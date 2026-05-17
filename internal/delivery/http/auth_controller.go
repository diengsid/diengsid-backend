package http

import (
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"id.diengs.backend/internal/model"
	"id.diengs.backend/internal/usecase"
)

type AuthController struct {
	AuthUseCase *usecase.AuthUseCase
	Log         *logrus.Logger
	Viper       *viper.Viper
}

func NewAuthController(authUseCase *usecase.AuthUseCase, log *logrus.Logger, cfg *viper.Viper) *AuthController {
	return &AuthController{
		AuthUseCase: authUseCase,
		Log:         log,
		Viper:       cfg,
	}
}

// setTokenCookie sets the auth cookie using values from config.
func (c *AuthController) setTokenCookie(ctx *fiber.Ctx, token string) {
	secure := c.Viper.GetBool("app.cookie_secure")
	domain := c.Viper.GetString("app.cookie_domain")
	sameSite := "Lax"
	if secure {
		sameSite = "None"
	}
	ctx.Cookie(&fiber.Cookie{
		Name:     "token",
		Value:    token,
		HTTPOnly: true,
		Secure:   secure,
		SameSite: sameSite,
		Domain:   domain,
		Path:     "/",
		MaxAge:   60 * 60 * 24 * 7, // 7 days
	})
}

// clearTokenCookie expires the auth cookie with the same attributes it was set with.
func (c *AuthController) clearTokenCookie(ctx *fiber.Ctx) {
	secure := c.Viper.GetBool("app.cookie_secure")
	domain := c.Viper.GetString("app.cookie_domain")
	sameSite := "Lax"
	if secure {
		sameSite = "None"
	}
	ctx.Cookie(&fiber.Cookie{
		Name:     "token",
		Value:    "",
		HTTPOnly: true,
		Secure:   secure,
		SameSite: sameSite,
		Domain:   domain,
		Path:     "/",
		MaxAge:   -1,
	})
}

// Send OTP
func (c *AuthController) SendOtp(ctx *fiber.Ctx) error {
	request := new(model.AuthSendOtpReq)
	if err := ctx.BodyParser(request); err != nil {
		c.Log.WithError(err).Error("failed to parse request body")
		return fiber.ErrBadRequest
	}

	if err := c.AuthUseCase.SendOtp(ctx.UserContext(), request); err != nil {
		c.Log.Error(err)
		return err
	}

	return ctx.JSON(model.WebResponse[string]{
		Success: true,
		Message: "success send otp",
	})
}

// Verify OTP
func (c *AuthController) VeriftOtp(ctx *fiber.Ctx) error {
	request := new(model.AuthVerifyOtpRequest)
	if err := ctx.BodyParser(request); err != nil {
		c.Log.WithError(err).Error("failed to parse request body")
		return fiber.ErrBadRequest
	}

	if err := c.AuthUseCase.VerifyOtp(ctx.UserContext(), request); err != nil {
		c.Log.Error(err)
		return err
	}

	return ctx.JSON(model.WebResponse[string]{
		Success: true,
		Message: "success verify otp",
	})
}

// Auth Google
func (c *AuthController) AuthGoogle(ctx *fiber.Ctx) error {
	userAgent := ctx.Get(fiber.HeaderUserAgent)
	ip := ctx.IP()

	request := new(model.AuthGoogleRequest)
	request.IP = ip
	request.UserAgent = userAgent

	if err := ctx.BodyParser(request); err != nil {
		c.Log.WithError(err).Error("failed to parse request body")
		return fiber.ErrBadRequest
	}

	response, err := c.AuthUseCase.AuthGoogle(ctx.UserContext(), request)
	if err != nil {
		c.Log.WithError(err).Error("failed to login user")
		return err
	}

	c.setTokenCookie(ctx, response.Token)

	return ctx.JSON(model.WebResponse[*model.AuthResponse]{
		Success: true,
		Message: "success",
		Data:    response,
	})
}

// Logout
func (c *AuthController) Logout(ctx *fiber.Ctx) error {
	token := ctx.Cookies("token")
	if token == "" {
		return fiber.ErrUnauthorized
	}

	if err := c.AuthUseCase.Logout(ctx.UserContext(), token); err != nil {
		c.Log.WithError(err).Error("failed to logout user")
		return err
	}

	c.clearTokenCookie(ctx)

	return ctx.JSON(model.WebResponse[any]{
		Message: "logout successfully",
		Success: true,
	})
}

// Register
func (c *AuthController) Register(ctx *fiber.Ctx) error {
	req := new(model.RegisterRequest)
	if err := ctx.BodyParser(req); err != nil {
		c.Log.WithError(err).Error("failed to parse register request body")
		return fiber.ErrBadRequest
	}

	response, err := c.AuthUseCase.Register(ctx.UserContext(), req)
	if err != nil {
		c.Log.WithError(err).Error("failed to register user")
		return err
	}

	c.setTokenCookie(ctx, response.Token)

	return ctx.Status(fiber.StatusCreated).JSON(model.WebResponse[*model.AuthResponse]{
		Success: true,
		Message: "register success",
		Data:    response,
	})
}

// Login
func (c *AuthController) Login(ctx *fiber.Ctx) error {
	req := new(model.LoginRequest)
	if err := ctx.BodyParser(req); err != nil {
		c.Log.WithError(err).Error("failed to parse login request body")
		return fiber.ErrBadRequest
	}

	response, err := c.AuthUseCase.Login(ctx.UserContext(), req)
	if err != nil {
		c.Log.WithError(err).Error("failed to login user")
		return err
	}

	c.setTokenCookie(ctx, response.Token)

	return ctx.JSON(model.WebResponse[*model.AuthResponse]{
		Success: true,
		Message: "login success",
		Data:    response,
	})
}

// Get Current
func (c *AuthController) Current(ctx *fiber.Ctx) error {
	user := ctx.Locals("user").(*model.UserResponse)
	return ctx.JSON(model.WebResponse[*model.UserResponse]{
		Data: user,
	})
}
