package usecase

import (
	"context"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"id.diengs.backend/internal/entity"
	"id.diengs.backend/internal/model"
	"id.diengs.backend/internal/repository"
)

type AvailabilityUseCase struct {
	DB               *gorm.DB
	Log              *logrus.Logger
	Validate         *validator.Validate
	AvailabilityRepo *repository.AvailabilityRepo
	RentableRepo     *repository.RentableRepo
}

func NewAvailabilityUseCase(
	db *gorm.DB,
	log *logrus.Logger,
	validate *validator.Validate,
	availabilityRepo *repository.AvailabilityRepo,
	rentableRepo *repository.RentableRepo,
) *AvailabilityUseCase {
	return &AvailabilityUseCase{
		DB:               db,
		Log:              log,
		Validate:         validate,
		AvailabilityRepo: availabilityRepo,
		RentableRepo:     rentableRepo,
	}
}

func (u *AvailabilityUseCase) Check(ctx context.Context, rentableID string, req *model.CheckAvailabilityRequest) ([]model.AvailabilityResponse, error) {
	tx := u.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	if err := u.Validate.Struct(req); err != nil {
		u.Log.WithError(err).Error("FAILED TO VALIDATE REQUEST.")
		return nil, fiber.ErrBadRequest
	}

	checkIn, err := time.Parse("2006-01-02", req.CheckIn)
	if err != nil {
		u.Log.WithError(err).Error("INVALID CHECK-IN DATE FORMAT.")
		return nil, fiber.ErrBadRequest
	}

	checkOut, err := time.Parse("2006-01-02", req.CheckOut)
	if err != nil {
		u.Log.WithError(err).Error("INVALID CHECK-OUT DATE FORMAT.")
		return nil, fiber.ErrBadRequest
	}

	if !checkOut.After(checkIn) {
		u.Log.Error("CHECK-OUT MUST BE AFTER CHECK-IN.")
		return nil, fiber.ErrBadRequest
	}

	rentable := new(entity.Rentable)
	if err := u.RentableRepo.FindById(tx, rentable, rentableID); err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fiber.ErrNotFound
		}
		u.Log.WithError(err).Error("FAILED TO GET RENTABLE.")
		return nil, fiber.ErrInternalServerError
	}

	var avails []entity.Availability
	if err := u.AvailabilityRepo.FindByRentableAndDateRange(tx, &avails, rentableID, checkIn, checkOut); err != nil {
		u.Log.WithError(err).Error("FAILED TO GET AVAILABILITY.")
		return nil, fiber.ErrInternalServerError
	}

	availMap := make(map[string]entity.Availability)
	for _, a := range avails {
		availMap[time.Unix(a.Date, 0).UTC().Format("2006-01-02")] = a
	}

	dates := datesInRange(checkIn, checkOut)
	responses := make([]model.AvailabilityResponse, len(dates))
	for i, date := range dates {
		dateStr := date.Format("2006-01-02")
		count := rentable.Stock
		var priceOverride *float64

		if a, exists := availMap[dateStr]; exists {
			count = a.AvailableCount
			priceOverride = a.PriceOverride
		}

		responses[i] = model.AvailabilityResponse{
			Date:           dateStr,
			AvailableCount: count,
			PriceOverride:  priceOverride,
			IsAvailable:    count > 0,
		}
	}

	if err := tx.Commit().Error; err != nil {
		u.Log.WithError(err).Error("FAILED TO COMMIT TRANSACTION.")
		return nil, fiber.ErrInternalServerError
	}

	return responses, nil
}
