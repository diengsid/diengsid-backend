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

type RentableUseCase struct {
	DB           *gorm.DB
	Log          *logrus.Logger
	Validate     *validator.Validate
	RentableRepo *repository.RentableRepo
	PropertyRepo *repository.PropertyRepo
}

func NewRentableUseCase(
	db *gorm.DB,
	log *logrus.Logger,
	validate *validator.Validate,
	rentableRepo *repository.RentableRepo,
	propertyRepo *repository.PropertyRepo,
) *RentableUseCase {
	return &RentableUseCase{
		DB:           db,
		Log:          log,
		Validate:     validate,
		RentableRepo: rentableRepo,
		PropertyRepo: propertyRepo,
	}
}

// create rentable
func (u *RentableUseCase) Create(ctx context.Context, request *model.RentableCreateRequest) (*model.RentableResponse, error) {
	tx := u.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	if err := u.Validate.Struct(request); err != nil {
		u.Log.WithError(err).Error("FAILED TO VALIDATE REQUEST.")
		return nil, fiber.ErrBadRequest
	}

	var propertyCount int64
	propertyCount, err := u.PropertyRepo.CountById(tx, request.PropertyID)
	if err != nil {
		u.Log.WithError(err).Error("FAILED TO VALIDATE PROPERTY ID.")
		return nil, fiber.ErrInternalServerError
	}
	if propertyCount == 0 {
		u.Log.Error("PROPERTY ID NOT FOUND.")
		return nil, fiber.ErrNotFound
	}

	if request.Type == "unit" && request.Stock > 1 {
		u.Log.Error("STOCK FOR UNIT TYPE CANNOT BE GREATER THAN 1.")
		return nil, fiber.ErrBadRequest
	}

	rentable := &entity.Rentable{
		PropertyID: request.PropertyID,
		Type:       request.Type,
		Name:       request.Name,
		ImageUrl:   request.ImageUrl,
		Capacity:   request.Capacity,
		BasePrice:  request.BasePrice,
		Discount:   request.Discount,
		Stock:      request.Stock,
	}

	if err := u.RentableRepo.Create(tx, rentable); err != nil {
		u.Log.WithError(err).Error("FAILED TO CREATE RENTABLE.")
		return nil, fiber.ErrInternalServerError
	}

	if err := tx.Commit().Error; err != nil {
		u.Log.WithError(err).Error("FAILED TO COMMIT TRANSACTION.")
		return nil, fiber.ErrInternalServerError
	}

	return model.RentableToResponse(rentable), nil
}
