package usecase

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/api/idtoken"
	"gorm.io/gorm"
	"id.diengs.backend/internal/entity"
	"id.diengs.backend/internal/lib"
	"id.diengs.backend/internal/model"
	"id.diengs.backend/internal/pkg"
	"id.diengs.backend/internal/pkg/mailview"
	"id.diengs.backend/internal/repository"
)

type AuthUseCase struct {
	DB           *gorm.DB
	Log          *logrus.Logger
	Validate     *validator.Validate
	Mail         *pkg.Mail
	UserRepo     *repository.UserRepo
	EmailOtpRepo *repository.EmailOtpRepo
	SessionRepo  *repository.SessionRepo
	Viper        *viper.Viper
}

func NewAuthUseCase(
	db *gorm.DB,
	log *logrus.Logger,
	validate *validator.Validate,
	mail *pkg.Mail,
	userRepo *repository.UserRepo,
	emailOtpRepo *repository.EmailOtpRepo,
	sessionRepo *repository.SessionRepo,
	viper *viper.Viper,
) *AuthUseCase {
	return &AuthUseCase{
		DB:           db,
		UserRepo:     userRepo,
		Log:          log,
		Validate:     validate,
		Mail:         mail,
		EmailOtpRepo: emailOtpRepo,
		SessionRepo:  sessionRepo,
		Viper:        viper,
	}
}

// Send Email OTP
func (u *AuthUseCase) SendOtp(ctx context.Context, request *model.AuthSendOtpReq) error {
	// transaction
	tx := u.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	// validate request
	if err := u.Validate.Struct(request); err != nil {
		u.Log.WithError(err).Error("FAILED TO VALIDATE REQUEST.")
		return fiber.ErrBadRequest
	}

	// create otp
	otp := rand.Intn(900000) + 100000 // 6 digit
	otpString := fmt.Sprint(otp)
	hashedOtp, _ := bcrypt.GenerateFromPassword([]byte(otpString), bcrypt.DefaultCost)

	// expired code
	expiredCode := time.Now().Add(time.Duration(5) * time.Minute).UnixMilli()

	// create to db
	emailOtp := &entity.EmailOtp{
		Email:     request.Email,
		OtpCode:   string(hashedOtp),
		ExpiredAt: expiredCode,
	}

	if err := u.EmailOtpRepo.Create(tx, emailOtp); err != nil {
		u.Log.WithError(err).Error("FAILED TO CREATE DB.")
		return fiber.ErrInternalServerError
	}

	// send mail
	bodyEmail := mailview.RegisterOtpMailView(otpString)
	err := u.Mail.SendMail([]string{request.Email}, "Kode pengamanan anda "+otpString, bodyEmail)
	if err != nil {
		u.Log.WithError(err).Error("FAILED TO SEND EMAIL.")
		return fiber.ErrInternalServerError
	}

	// commit transaction
	if err := tx.Commit().Error; err != nil {
		u.Log.WithError(err).Error("FAILED TO COMMIT TRANSACTION.")
		return fiber.ErrInternalServerError
	}

	return nil
}

// Verify Email OTP
func (u *AuthUseCase) VerifyOtp(ctx context.Context, request *model.AuthVerifyOtpRequest) error {
	// transaction
	tx := u.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	// validate request
	if err := u.Validate.Struct(request); err != nil {
		u.Log.WithError(err).Error("FAILED TO VALIDATE REQUEST.")
		return fiber.ErrBadRequest
	}

	// find OTP
	emailOtp := new(entity.EmailOtp)
	err := u.EmailOtpRepo.FindActiveAndEmail(tx, emailOtp, request.Email)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return fiber.ErrNotFound
		}
		u.Log.WithError(err).Error("FAILED TO FIND EMAIL OTP.")
		return fiber.ErrInternalServerError
	}

	if emailOtp.AttemptCount >= emailOtp.MaxAttempt {
		return fiber.NewError(fiber.StatusTooManyRequests, "TO MANY ATTEMPS.")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(emailOtp.OtpCode), []byte(request.Otp)); err != nil {
		u.Log.WithError(err).Error("FAILED TO FIND EMAIL OTP.")
		return fiber.ErrNotFound
	}

	// Update OTP
	emailOtp.IsUsed = true
	if err := u.EmailOtpRepo.Update(tx, emailOtp); err != nil {
		u.Log.WithError(err).Error("FAILED TO UPDATE EMAIL OTP.")
		return fiber.ErrInternalServerError
	}

	// commit transaction
	if err := tx.Commit().Error; err != nil {
		u.Log.WithError(err).Error("FAILED TO COMMIT TRANSACTION.")
		return fiber.ErrInternalServerError
	}

	return nil
}

// Auth with google
func (u *AuthUseCase) AuthGoogle(ctx context.Context, request *model.AuthGoogleRequest) (*model.AuthResponse, error) {
	clientId := u.Viper.GetString("google.clientId")
	// transaction
	tx := u.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	// validate request
	// validate request
	if err := u.Validate.Struct(request); err != nil {
		u.Log.WithError(err).Error("FAILED TO VALIDATE REQUEST.")
		return nil, fiber.ErrBadRequest
	}

	// validate token
	payload, err := idtoken.Validate(ctx, request.Token, clientId)
	if err != nil {
		return nil, err
	}

	email := payload.Claims["email"].(string)
	name := payload.Claims["name"].(string)
	picture := payload.Claims["picture"].(string)
	provider := "google"
	providerID := payload.Claims["sub"].(string)

	// find user
	user := new(entity.User)
	err = u.UserRepo.FindByEmail(tx, user, email)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// user belum ada → register
			user = &entity.User{
				Email:         email,
				Name:          name,
				Picture:       &picture,
				Provider:      &provider,
				ProviderID:    &providerID,
				EmailVerified: true,
				Role:          "USER",
			}
			err = u.UserRepo.Create(tx, user)
			if err != nil {
				u.Log.WithError(err).Error("FAILED TO CREATE USER.")
				return nil, fiber.ErrInternalServerError
			}

		} else {
			u.Log.WithError(err).Error("FAILED TO FIND EMAIL.")
			return nil, fiber.ErrInternalServerError
		}
	}

	// create token
	token, err := lib.GenerateToken(32)
	if err != nil {
		return nil, fiber.ErrInternalServerError
	}

	// create session
	session := &entity.Session{
		UserID:    user.ID,
		Token:     token,
		IPAddress: &request.IP,
		UserAgent: &request.UserAgent,
		ExpiredAt: time.Now().Add(1000 * time.Hour).UnixMilli(),
	}

	if err := u.SessionRepo.Create(tx, session); err != nil {
		u.Log.WithError(err).Error("Failed to create session")
		return nil, fiber.ErrInternalServerError
	}

	u.Log.Println("TES")

	if err := tx.Commit().Error; err != nil {
		u.Log.Warnf("Failed commit transaction : %+v", err)
		return nil, fiber.ErrInternalServerError
	}

	return &model.AuthResponse{
		User:  *model.UserToResponse(user),
		Token: token,
	}, nil
}

// Auth logout
func (c *AuthUseCase) Logout(ctx context.Context, token string) error {
	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	// delete session by user id
	if err := c.SessionRepo.DeleteByToken(tx, token); err != nil {
		c.Log.WithError(err).Error("Failed to delete session by token")
		return fiber.ErrInternalServerError
	}

	if err := tx.Commit().Error; err != nil {
		c.Log.Warnf("Failed commit transaction : %+v", err)
		return fiber.ErrInternalServerError
	}

	return nil
}

// Verify User
func (c *AuthUseCase) Verify(ctx context.Context, request *model.VerifyUserRequest) (*model.UserResponse, error) {
	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	// find session
	session := new(entity.Session)
	if err := c.SessionRepo.FindByToken(tx, session, request.Token); err != nil {
		c.Log.Warnf("Failed find user by token : %+v", err)
		return nil, fiber.ErrUnauthorized
	}

	expiredAt := time.Unix(session.CreatedAt, 0)

	// Check expiry
	if expiredAt.Before(time.Now()) {
		if err := c.SessionRepo.Delete(tx, session); err != nil {
			c.Log.WithError(err).Error("Failed to delete session by user id")
			return nil, fiber.ErrInternalServerError
		}
		return nil, fiber.NewError(fiber.StatusUnauthorized, "Session expired")
	}

	// find user
	user := new(entity.User)
	if err := c.UserRepo.FindById(tx, user, session.UserID); err != nil {
		c.Log.Warnf("Failed find user by token : %+v", err)
		return nil, fiber.ErrUnauthorized
	}

	if err := tx.Commit().Error; err != nil {
		c.Log.Warnf("Failed commit transaction : %+v", err)
		return nil, fiber.ErrInternalServerError
	}

	return model.UserToResponse(user), nil
}

func (u *AuthUseCase) Register(ctx context.Context, req *model.RegisterRequest) (*model.AuthResponse, error) {
	tx := u.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	if err := u.Validate.Struct(req); err != nil {
		u.Log.WithError(err).Error("FAILED TO VALIDATE REGISTER REQUEST.")
		return nil, fiber.ErrBadRequest
	}

	existing := new(entity.User)
	err := u.UserRepo.FindByEmail(tx, existing, req.Email)
	if err == nil {
		return nil, fiber.NewError(fiber.StatusConflict, "email already registered")
	}
	if err != gorm.ErrRecordNotFound {
		u.Log.WithError(err).Error("FAILED TO CHECK EMAIL.")
		return nil, fiber.ErrInternalServerError
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		u.Log.WithError(err).Error("FAILED TO HASH PASSWORD.")
		return nil, fiber.ErrInternalServerError
	}

	user := &entity.User{
		Name:          req.Name,
		Email:         req.Email,
		Password:      string(hashed),
		EmailVerified: false,
		Role:          "USER",
	}

	if err := u.UserRepo.Create(tx, user); err != nil {
		u.Log.WithError(err).Error("FAILED TO CREATE USER.")
		return nil, fiber.ErrInternalServerError
	}

	token, err := lib.GenerateToken(32)
	if err != nil {
		u.Log.WithError(err).Error("FAILED TO GENERATE TOKEN.")
		return nil, fiber.ErrInternalServerError
	}

	session := &entity.Session{
		UserID:    user.ID,
		Token:     token,
		ExpiredAt: time.Now().Add(1000 * time.Hour).UnixMilli(),
	}

	if err := u.SessionRepo.Create(tx, session); err != nil {
		u.Log.WithError(err).Error("FAILED TO CREATE SESSION.")
		return nil, fiber.ErrInternalServerError
	}

	if err := tx.Commit().Error; err != nil {
		u.Log.WithError(err).Error("FAILED TO COMMIT TRANSACTION.")
		return nil, fiber.ErrInternalServerError
	}

	return &model.AuthResponse{
		User:  *model.UserToResponse(user),
		Token: token,
	}, nil
}

func (u *AuthUseCase) Login(ctx context.Context, req *model.LoginRequest) (*model.AuthResponse, error) {
	tx := u.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	if err := u.Validate.Struct(req); err != nil {
		u.Log.WithError(err).Error("FAILED TO VALIDATE LOGIN REQUEST.")
		return nil, fiber.ErrBadRequest
	}

	user := new(entity.User)
	if err := u.UserRepo.FindByEmail(tx, user, req.Email); err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fiber.ErrUnauthorized
		}
		u.Log.WithError(err).Error("FAILED TO FIND USER.")
		return nil, fiber.ErrInternalServerError
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, fiber.ErrUnauthorized
	}

	token, err := lib.GenerateToken(32)
	if err != nil {
		u.Log.WithError(err).Error("FAILED TO GENERATE TOKEN.")
		return nil, fiber.ErrInternalServerError
	}

	session := &entity.Session{
		UserID:    user.ID,
		Token:     token,
		ExpiredAt: time.Now().Add(24 * time.Hour).UnixMilli(),
	}

	if err := u.SessionRepo.Create(tx, session); err != nil {
		u.Log.WithError(err).Error("FAILED TO CREATE SESSION.")
		return nil, fiber.ErrInternalServerError
	}

	if err := tx.Commit().Error; err != nil {
		u.Log.WithError(err).Error("FAILED TO COMMIT TRANSACTION.")
		return nil, fiber.ErrInternalServerError
	}

	return &model.AuthResponse{
		User:  *model.UserToResponse(user),
		Token: token,
	}, nil
}
