package config

import (
	"gorm.io/gorm"
	"id.diengs.backend/internal/delivery/http"
	"id.diengs.backend/internal/delivery/http/route"
	"id.diengs.backend/internal/delivery/middleware"
	"id.diengs.backend/internal/pkg"
	"id.diengs.backend/internal/repository"
	"id.diengs.backend/internal/usecase"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type BootstrapConfig struct {
	DB       *gorm.DB
	App      *fiber.App
	Log      *logrus.Logger
	Validate *validator.Validate
	Config   *viper.Viper
	Mail     *pkg.Mail
}

func Bootstrap(cfg *BootstrapConfig) {
	// Repository Config
	userRepo := repository.NewUserRepo(cfg.Log)
	emailOtpRepo := repository.NewEmailOtpRepo(cfg.Log)
	sessionRepo := repository.NewSessionRepo(cfg.Log)
	propertyRepo := repository.NewPropertyRepo(cfg.Log)
	propertyImageRepo := repository.NewPropertyImageRepo(cfg.Log)
	hostProfileRepo := repository.NewHostProfileRepo(cfg.Log)
	rentableRepo := repository.NewRentableRepo(cfg.Log)
	bookingRepo := repository.NewBookingRepo(cfg.Log)
	availabilityRepo := repository.NewAvailabilityRepo(cfg.Log)
	paymentRepo := repository.NewPaymentRepo(cfg.Log)
	amenityRepo := repository.NewAmenityRepo(cfg.Log)
	attractionRepo := repository.NewTouristAttractionRepo(cfg.Log)

	// Use Case Config
	healthUseCase := usecase.NewHealthUseCase(cfg.Config)
	authUseCase := usecase.NewAuthUseCase(cfg.DB, cfg.Log, cfg.Validate, cfg.Mail, userRepo, emailOtpRepo, sessionRepo, cfg.Config)
	propertyUseCase := usecase.NewPropertyUseCase(cfg.DB, cfg.Log, cfg.Validate, propertyRepo, propertyImageRepo, hostProfileRepo)
	uploadUseCase := usecase.NewUploadUseCase(cfg.Log, cfg.Validate, cfg.Config)
	rentableUseCase := usecase.NewRentableUseCase(cfg.DB, cfg.Log, cfg.Validate, rentableRepo, propertyRepo)
	availabilityUseCase := usecase.NewAvailabilityUseCase(cfg.DB, cfg.Log, cfg.Validate, availabilityRepo, rentableRepo)
	bookingUseCase := usecase.NewBookingUseCase(cfg.DB, cfg.Log, cfg.Validate, bookingRepo, rentableRepo, availabilityRepo, propertyRepo, hostProfileRepo)
	amenityUseCase := usecase.NewAmenityUseCase(cfg.DB, cfg.Log, cfg.Validate, amenityRepo, propertyRepo, rentableRepo)
	attractionUseCase := usecase.NewTouristAttractionUseCase(cfg.DB, cfg.Log, cfg.Validate, attractionRepo, propertyRepo)

	dokuClient := pkg.NewDokuClient(cfg.Config)
	paymentUseCase := usecase.NewPaymentUseCase(cfg.DB, cfg.Log, bookingRepo, userRepo, paymentRepo, dokuClient)

	// Controller Config
	healthController := http.NewHealthController(healthUseCase, cfg.Log)
	authController := http.NewAuthController(authUseCase, cfg.Log, cfg.Config)
	propertyCotroller := http.NewPropertyController(cfg.Log, propertyUseCase)
	uploadController := http.NewUploadController(cfg.Log, uploadUseCase)
	rentableController := http.NewRentableController(cfg.Log, rentableUseCase)
	bookingController := http.NewBookingController(cfg.Log, bookingUseCase)
	availabilityController := http.NewAvailabilityController(cfg.Log, availabilityUseCase)
	paymentController := http.NewPaymentController(cfg.Log, paymentUseCase, dokuClient)
	amenityController := http.NewAmenityController(cfg.Log, amenityUseCase)
	attractionController := http.NewTouristAttractionController(cfg.Log, attractionUseCase)

	// setup middleware
	authMiddleware := middleware.NewAuth(authUseCase)
	adminMiddleware := middleware.NewAdmin()

	route.RouteConfig{
		App:                    cfg.App,
		AuthMiddleware:         authMiddleware,
		AdminMiddleware:        adminMiddleware,
		HealthController:       healthController,
		AuthController:         authController,
		PropertyController:     propertyCotroller,
		UploadController:       uploadController,
		RentableController:     rentableController,
		BookingController:      bookingController,
		AvailabilityController: availabilityController,
		PaymentController:      paymentController,
		AmenityController:           amenityController,
		TouristAttractionController: attractionController,
	}.Setup()
}
