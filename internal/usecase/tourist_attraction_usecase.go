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

type TouristAttractionUseCase struct {
	DB                   *gorm.DB
	Log                  *logrus.Logger
	Validate             *validator.Validate
	AttractionRepo       *repository.TouristAttractionRepo
	PropertyRepo         *repository.PropertyRepo
}

func NewTouristAttractionUseCase(
	db *gorm.DB,
	log *logrus.Logger,
	validate *validator.Validate,
	attractionRepo *repository.TouristAttractionRepo,
	propertyRepo *repository.PropertyRepo,
) *TouristAttractionUseCase {
	return &TouristAttractionUseCase{
		DB:             db,
		Log:            log,
		Validate:       validate,
		AttractionRepo: attractionRepo,
		PropertyRepo:   propertyRepo,
	}
}

// List semua tourist attractions
func (u *TouristAttractionUseCase) List(ctx context.Context) ([]model.TouristAttractionResponse, error) {
	var list []entity.TouristAttraction
	if err := u.AttractionRepo.FindAll(u.DB.WithContext(ctx), &list); err != nil {
		u.Log.WithError(err).Error("failed to list tourist attractions")
		return nil, fiber.ErrInternalServerError
	}

	out := make([]model.TouristAttractionResponse, len(list))
	for i := range list {
		out[i] = model.TouristAttractionToResponse(&list[i])
	}
	return out, nil
}

// Create tourist attraction baru
func (u *TouristAttractionUseCase) Create(ctx context.Context, req *model.TouristAttractionCreateRequest) (*model.TouristAttractionResponse, error) {
	if err := u.Validate.Struct(req); err != nil {
		return nil, fiber.ErrBadRequest
	}

	attraction := &entity.TouristAttraction{
		Name:        req.Name,
		Slug:        req.Slug,
		Description: req.Description,
		Address:     req.Address,
		Latitude:    req.Latitude,
		Longitude:   req.Longitude,
		Category:    req.Category,
		ImageURL:    req.ImageURL,
	}

	if err := u.AttractionRepo.Create(u.DB.WithContext(ctx), attraction); err != nil {
		u.Log.WithError(err).Error("failed to create tourist attraction")
		return nil, fiber.ErrInternalServerError
	}

	resp := model.TouristAttractionToResponse(attraction)
	return &resp, nil
}

// GetNearbyByPropertyID mengambil daftar objek wisata terdekat dari sebuah property
func (u *TouristAttractionUseCase) GetNearbyByPropertyID(ctx context.Context, propertyID string) ([]model.NearbyAttractionResponse, error) {
	db := u.DB.WithContext(ctx)

	var count int64
	if err := db.Model(&entity.Property{}).Where("id = ?", propertyID).Count(&count).Error; err != nil {
		return nil, fiber.ErrInternalServerError
	}
	if count == 0 {
		return nil, fiber.ErrNotFound
	}

	var list []entity.PropertyNearbyAttraction
	if err := u.AttractionRepo.FindNearbyByPropertyID(db, &list, propertyID); err != nil {
		u.Log.WithError(err).Error("failed to get nearby attractions")
		return nil, fiber.ErrInternalServerError
	}

	return model.NearbyAttractionsToResponse(list), nil
}

// SetNearbyAttractions mengganti seluruh daftar objek wisata terdekat dari sebuah property
func (u *TouristAttractionUseCase) SetNearbyAttractions(ctx context.Context, propertyID string, req *model.SetNearbyAttractionsRequest) ([]model.NearbyAttractionResponse, error) {
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

	tx := db.Begin()
	defer tx.Rollback()

	if err := tx.Exec("DELETE FROM property_nearby_attractions WHERE property_id = ?", propertyID).Error; err != nil {
		u.Log.WithError(err).Error("failed to delete old nearby attractions")
		return nil, fiber.ErrInternalServerError
	}

	for _, item := range req.Attractions {
		row := &entity.PropertyNearbyAttraction{
			PropertyID:          propertyID,
			TouristAttractionID: item.TouristAttractionID,
			DistanceKm:          item.DistanceKm,
			DurationMinutes:     item.DurationMinutes,
			SortOrder:           item.SortOrder,
		}
		if err := tx.Create(row).Error; err != nil {
			u.Log.WithError(err).Error("failed to insert nearby attraction")
			return nil, fiber.ErrInternalServerError
		}
	}

	if err := tx.Commit().Error; err != nil {
		u.Log.WithError(err).Error("failed to commit set nearby attractions")
		return nil, fiber.ErrInternalServerError
	}

	var list []entity.PropertyNearbyAttraction
	if err := u.AttractionRepo.FindNearbyByPropertyID(db, &list, propertyID); err != nil {
		return nil, fiber.ErrInternalServerError
	}

	return model.NearbyAttractionsToResponse(list), nil
}
