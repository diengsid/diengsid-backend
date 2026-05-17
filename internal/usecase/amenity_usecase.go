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

type AmenityUseCase struct {
	DB           *gorm.DB
	Log          *logrus.Logger
	Validate     *validator.Validate
	AmenityRepo  *repository.AmenityRepo
	PropertyRepo *repository.PropertyRepo
	RentableRepo *repository.RentableRepo
}

func NewAmenityUseCase(
	db *gorm.DB,
	log *logrus.Logger,
	validate *validator.Validate,
	amenityRepo *repository.AmenityRepo,
	propertyRepo *repository.PropertyRepo,
	rentableRepo *repository.RentableRepo,
) *AmenityUseCase {
	return &AmenityUseCase{
		DB:           db,
		Log:          log,
		Validate:     validate,
		AmenityRepo:  amenityRepo,
		PropertyRepo: propertyRepo,
		RentableRepo: rentableRepo,
	}
}

func (u *AmenityUseCase) List(ctx context.Context) ([]model.AmenityResponse, error) {
	var amenities []entity.Amenity
	if err := u.AmenityRepo.FindAll(u.DB.WithContext(ctx), &amenities); err != nil {
		u.Log.WithError(err).Error("failed to list amenities")
		return nil, fiber.ErrInternalServerError
	}
	return model.AmenitiesToResponse(amenities), nil
}

func (u *AmenityUseCase) Create(ctx context.Context, req *model.AmenityCreateRequest) (*model.AmenityResponse, error) {
	if err := u.Validate.Struct(req); err != nil {
		return nil, fiber.ErrBadRequest
	}

	amenity := &entity.Amenity{
		Name:     req.Name,
		Icon:     req.Icon,
		Category: req.Category,
	}
	if err := u.AmenityRepo.Create(u.DB.WithContext(ctx), amenity); err != nil {
		u.Log.WithError(err).Error("failed to create amenity")
		return nil, fiber.ErrInternalServerError
	}

	resp := model.AmenityToResponse(amenity)
	return &resp, nil
}

// SetPropertyAmenities replaces the full set of amenities for a property.
func (u *AmenityUseCase) SetPropertyAmenities(ctx context.Context, propertyID string, req *model.SetAmenitiesRequest) ([]model.AmenityResponse, error) {
	if err := u.Validate.Struct(req); err != nil {
		return nil, fiber.ErrBadRequest
	}

	db := u.DB.WithContext(ctx)
	var count int64
	if err := db.Model(&entity.Property{}).Where("id = ?", propertyID).Count(&count).Error; err != nil {
		return nil, fiber.ErrInternalServerError
	}
	if count == 0 {
		return nil, fiber.ErrNotFound
	}

	var amenities []entity.Amenity
	if len(req.AmenityIDs) > 0 {
		if err := u.AmenityRepo.FindByIDs(db, &amenities, req.AmenityIDs); err != nil {
			return nil, fiber.ErrInternalServerError
		}
	}

	if err := replaceJoinTable(db, "property_amenities", "property_id", propertyID, req.AmenityIDs); err != nil {
		u.Log.WithError(err).Error("failed to set property amenities")
		return nil, fiber.ErrInternalServerError
	}

	return model.AmenitiesToResponse(amenities), nil
}

// SetRentableAmenities replaces the full set of amenities for a rentable.
func (u *AmenityUseCase) SetRentableAmenities(ctx context.Context, rentableID string, req *model.SetAmenitiesRequest) ([]model.AmenityResponse, error) {
	if err := u.Validate.Struct(req); err != nil {
		return nil, fiber.ErrBadRequest
	}

	db := u.DB.WithContext(ctx)
	var count int64
	if err := db.Model(&entity.Rentable{}).Where("id = ?", rentableID).Count(&count).Error; err != nil {
		return nil, fiber.ErrInternalServerError
	}
	if count == 0 {
		return nil, fiber.ErrNotFound
	}

	var amenities []entity.Amenity
	if len(req.AmenityIDs) > 0 {
		if err := u.AmenityRepo.FindByIDs(db, &amenities, req.AmenityIDs); err != nil {
			return nil, fiber.ErrInternalServerError
		}
	}

	if err := replaceJoinTable(db, "rentable_amenities", "rentable_id", rentableID, req.AmenityIDs); err != nil {
		u.Log.WithError(err).Error("failed to set rentable amenities")
		return nil, fiber.ErrInternalServerError
	}

	return model.AmenitiesToResponse(amenities), nil
}

// replaceJoinTable deletes all rows for ownerID then inserts the new amenity_ids.
func replaceJoinTable(db *gorm.DB, table, ownerCol, ownerID string, amenityIDs []string) error {
	if err := db.Exec("DELETE FROM "+table+" WHERE "+ownerCol+" = ?", ownerID).Error; err != nil {
		return err
	}
	for _, aid := range amenityIDs {
		if err := db.Exec(
			"INSERT INTO "+table+" ("+ownerCol+", amenity_id) VALUES (?, ?) ON CONFLICT DO NOTHING",
			ownerID, aid,
		).Error; err != nil {
			return err
		}
	}
	return nil
}
