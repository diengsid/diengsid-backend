package usecase

import (
	"context"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"id.diengs.backend/internal/entity"
	"id.diengs.backend/internal/model"
	"id.diengs.backend/internal/repository"
)

type PropertyUseCase struct {
	DB              *gorm.DB
	Log             *logrus.Logger
	Validate        *validator.Validate
	ProperyRepo     *repository.PropertyRepo
	HostProfileRepo *repository.HostProfileRepo
}

func NewPropertyUseCase(
	db *gorm.DB,
	log *logrus.Logger,
	validate *validator.Validate,
	propertyRepo *repository.PropertyRepo,
	hostProfileRepo *repository.HostProfileRepo,
) *PropertyUseCase {
	return &PropertyUseCase{
		DB:              db,
		Log:             log,
		Validate:        validate,
		ProperyRepo:     propertyRepo,
		HostProfileRepo: hostProfileRepo,
	}
}

// Create Property
func (u *PropertyUseCase) Create(ctx context.Context, req *model.PropertyCreateRequest) (*model.PropertyResponse, error) {
	tx := u.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	// Validasi input
	if err := u.Validate.Struct(req); err != nil {
		u.Log.WithError(err).Error("FAILED TO VALIDATE REQUEST.")
		return nil, fiber.ErrBadRequest
	}

	// validasi jika HostID tidak disediakan, maka Host harus disediakan
	if req.HostID == nil && req.Host == nil {
		u.Log.Error("HOST ID OR HOST DATA MUST BE PROVIDED.")
		return nil, fiber.ErrBadRequest
	}

	hostProfileId := req.HostID

	// find host by id
	hostProfile := new(entity.HostProfile)
	if req.HostID != nil {
		var err error
		err = u.HostProfileRepo.FindById(tx, hostProfile, *req.HostID)
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				u.Log.WithError(err).Error("HOST NOT FOUND.")
				return nil, fiber.ErrNotFound
			}
			u.Log.WithError(err).Error("FAILED TO GET HOST.")
			return nil, fiber.ErrInternalServerError
		}
	}

	// Jika HostID tidak disediakan, buat Host baru
	if hostProfile.ID == "" {
		// Logika untuk membuat host baru
		host := &entity.HostProfile{
			Name:              req.Host.Name,
			Email:             req.Host.Email,
			PhoneNumber:       req.Host.PhoneNumber,
			ProfilePictureURL: req.Host.ProfilePictureURL,
			Address:           req.Host.Address,
			BankAccountName:   req.Host.BankAccountName,
			BankAccountNumber: req.Host.BankAccountNumber,
			KTPNumber:         req.Host.KTPNumber,
			Bio:               req.Host.Bio,
		}

		if err := u.HostProfileRepo.Create(tx, host); err != nil {
			u.Log.WithError(err).Error("FAILED TO CREATE HOST.")
			return nil, fiber.ErrInternalServerError
		}
		hostProfileId = &host.ID
	}

	property := &entity.Property{
		ExperienceID: req.ExperienceID,
		HostID:       *hostProfileId,
		PropertyType: req.PropertyType,
		BookingType:  req.BookingType,
	}

	if err := u.ProperyRepo.Create(tx, property); err != nil {
		u.Log.WithError(err).Error("FAILED TO CREATE PROPERTY.")
		return nil, fiber.ErrInternalServerError
	}

	result := new(entity.Property)
	err := u.ProperyRepo.FindById(tx, result, property.ID, "Host", "Experience")
	if err != nil {
		return nil, fiber.ErrInternalServerError
	}

	if err := tx.Commit().Error; err != nil {
		u.Log.WithError(err).Error("FAILED TO COMMIT TRANSACTION.")
		return nil, fiber.ErrInternalServerError
	}

	return model.PropertyToResponse(result), nil
}

// Get Property By ID
func (u *PropertyUseCase) GetByID(ctx context.Context, id string) (*model.PropertyResponse, error) {
	tx := u.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	property := new(entity.Property)
	err := u.ProperyRepo.FindByExperienceID(tx, property, id, "Host", "Experience", "Experience.Images", "Rentable")
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			u.Log.WithError(err).Error("PROPERTY NOT FOUND.")
			return nil, fiber.ErrNotFound
		}
		u.Log.WithError(err).Error("FAILED TO GET PROPERTY.")
		return nil, fiber.ErrInternalServerError
	}

	if err := tx.Commit().Error; err != nil {
		u.Log.WithError(err).Error("FAILED TO COMMIT TRANSACTION.")
		return nil, fiber.ErrInternalServerError
	}

	return model.PropertyToResponse(property), nil
}
