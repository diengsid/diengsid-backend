package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"id.diengs.backend/internal/entity"
	"id.diengs.backend/internal/model"
	"id.diengs.backend/internal/pkg"
	"id.diengs.backend/internal/repository"
)

type PropertyUseCase struct {
	DB                *gorm.DB
	Log               *logrus.Logger
	Validate          *validator.Validate
	ProperyRepo       *repository.PropertyRepo
	PropertyImageRepo *repository.PropertyImageRepo
	HostProfileRepo   *repository.HostProfileRepo
	Fonnte            *pkg.FonnteClient
}

func NewPropertyUseCase(
	db *gorm.DB,
	log *logrus.Logger,
	validate *validator.Validate,
	propertyRepo *repository.PropertyRepo,
	propertyImageRepo *repository.PropertyImageRepo,
	hostProfileRepo *repository.HostProfileRepo,
	fonnte *pkg.FonnteClient,
) *PropertyUseCase {
	return &PropertyUseCase{
		DB:                db,
		Log:               log,
		Validate:          validate,
		ProperyRepo:       propertyRepo,
		PropertyImageRepo: propertyImageRepo,
		HostProfileRepo:   hostProfileRepo,
		Fonnte:            fonnte,
	}
}

// Create Property
func (u *PropertyUseCase) Create(ctx context.Context, req *model.PropertyCreateRequest) (*model.PropertyResponse, error) {
	tx := u.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	if err := u.Validate.Struct(req); err != nil {
		u.Log.WithError(err).Error("FAILED TO VALIDATE REQUEST.")
		return nil, fiber.ErrBadRequest
	}

	if req.HostID == nil && req.Host == nil {
		u.Log.Error("HOST ID OR HOST DATA MUST BE PROVIDED.")
		return nil, fiber.ErrBadRequest
	}

	hostProfileId := req.HostID

	hostProfile := new(entity.HostProfile)
	if req.HostID != nil {
		if err := u.HostProfileRepo.FindById(tx, hostProfile, *req.HostID); err != nil {
			if err == gorm.ErrRecordNotFound {
				u.Log.WithError(err).Error("HOST NOT FOUND.")
				return nil, fiber.ErrNotFound
			}
			u.Log.WithError(err).Error("FAILED TO GET HOST.")
			return nil, fiber.ErrInternalServerError
		}
	}

	if hostProfile.ID == "" {
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
		HostID:       *hostProfileId,
		PropertyType: req.PropertyType,
		BookingType:  req.BookingType,
		Title:        req.Title,
		Address:      req.Address,
		Description:  req.Description,
		ThumbnailURL: req.ThumbnailURL,
		Lat:          req.Lat,
		Lng:          req.Lng,
	}

	if err := u.ProperyRepo.Create(tx, property); err != nil {
		u.Log.WithError(err).Error("FAILED TO CREATE PROPERTY.")
		return nil, fiber.ErrInternalServerError
	}

	for _, img := range req.Images {
		propertyImage := &entity.PropertyImage{
			PropertyID: property.ID,
			ImageURL:   img.ImageURL,
			IsPrimary:  img.IsPrimary,
		}
		if err := u.PropertyImageRepo.Create(tx, propertyImage); err != nil {
			u.Log.WithError(err).Error("FAILED TO CREATE PROPERTY IMAGE.")
			return nil, fiber.ErrInternalServerError
		}
		property.Images = append(property.Images, *propertyImage)
	}

	for _, amenityID := range req.AmenityIDs {
		if err := tx.Exec(
			"INSERT INTO property_amenities (property_id, amenity_id) VALUES (?, ?) ON CONFLICT DO NOTHING",
			property.ID, amenityID,
		).Error; err != nil {
			u.Log.WithError(err).Error("FAILED TO ASSOCIATE AMENITY.")
			return nil, fiber.ErrInternalServerError
		}
	}

	result := new(entity.Property)
	if err := u.ProperyRepo.FindById(tx, result, property.ID, "Host", "Images", "Amenities"); err != nil {
		return nil, fiber.ErrInternalServerError
	}

	if err := tx.Commit().Error; err != nil {
		u.Log.WithError(err).Error("FAILED TO COMMIT TRANSACTION.")
		return nil, fiber.ErrInternalServerError
	}

	// Send WhatsApp notification to host (non-blocking)
	go func() {
		phone := result.Host.PhoneNumber
		if phone == "" || u.Fonnte == nil {
			return
		}
		msg := fmt.Sprintf(
			"Halo %s!\n\nProperti baru Anda telah berhasil terdaftar di Diengs.id.\n\nDetail Properti:\nNama    : %s\nTipe    : %s\nAlamat  : %s\n\nSilakan kelola properti Anda melalui dashboard admin.\n\nTerima kasih,\nTim Diengs.id",
			result.Host.Name,
			result.Title,
			result.PropertyType,
			result.Address,
		)
		if _, err := u.Fonnte.SendOne(phone, msg); err != nil {
			u.Log.WithError(err).Warn("failed to send whatsapp notification to host")
		}
	}()

	return model.PropertyToResponse(result), nil
}

// Search Properties
func (u *PropertyUseCase) Search(ctx context.Context, req *model.SearchPropertyRequest) ([]model.PropertyResponse, int64, error) {
	if err := u.Validate.Struct(req); err != nil {
		u.Log.WithError(err).Error("FAILED TO VALIDATE SEARCH REQUEST.")
		return nil, 0, fiber.ErrBadRequest
	}

	var checkInUnix, checkOutUnix int64

	if req.CheckIn != "" && req.CheckOut != "" {
		checkIn, err := time.Parse("2006-01-02", req.CheckIn)
		if err != nil {
			u.Log.WithError(err).Error("INVALID CHECK_IN FORMAT.")
			return nil, 0, fiber.ErrBadRequest
		}
		checkOut, err := time.Parse("2006-01-02", req.CheckOut)
		if err != nil {
			u.Log.WithError(err).Error("INVALID CHECK_OUT FORMAT.")
			return nil, 0, fiber.ErrBadRequest
		}
		if !checkOut.After(checkIn) {
			u.Log.Error("CHECK_OUT MUST BE AFTER CHECK_IN.")
			return nil, 0, fiber.ErrBadRequest
		}
		checkInUnix = checkIn.Unix()
		checkOutUnix = checkOut.Unix()
	}

	guestCount := req.GuestCount
	if guestCount < 1 {
		guestCount = 1
	}

	properties, total, err := u.ProperyRepo.Search(u.DB.WithContext(ctx), req, checkInUnix, checkOutUnix, guestCount)
	if err != nil {
		u.Log.WithError(err).Error("FAILED TO SEARCH PROPERTIES.")
		return nil, 0, fiber.ErrInternalServerError
	}

	responses := make([]model.PropertyResponse, len(properties))
	for i := range properties {
		responses[i] = *model.PropertyToResponse(&properties[i])
	}

	return responses, total, nil
}

// Get Property By ID
func (u *PropertyUseCase) GetByID(ctx context.Context, id string) (*model.PropertyResponse, error) {
	tx := u.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	property := new(entity.Property)
	err := u.ProperyRepo.FindById(tx, property, id, "Host", "Images", "Rentable", "Rentable.Amenities", "Amenities", "NearbyAttractions", "NearbyAttractions.TouristAttraction")
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
