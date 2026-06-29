package usecase

import (
	"context"
	"math"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"id.diengs.backend/internal/entity"
	"id.diengs.backend/internal/model"
	"id.diengs.backend/internal/repository"
)

// haversineKm returns the great-circle distance in kilometres between two lat/lng points.
func haversineKm(lat1, lon1, lat2, lon2 float64) float64 {
	const R = 6371.0
	dLat := (lat2 - lat1) * math.Pi / 180
	dLon := (lon2 - lon1) * math.Pi / 180
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1*math.Pi/180)*math.Cos(lat2*math.Pi/180)*
			math.Sin(dLon/2)*math.Sin(dLon/2)
	return R * 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
}

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

// SetNearbyAttractions mengganti seluruh daftar objek wisata terdekat dari sebuah property.
// DistanceKm dan DurationMinutes dihitung otomatis dari koordinat masing-masing.
func (u *TouristAttractionUseCase) SetNearbyAttractions(ctx context.Context, propertyID string, req *model.SetNearbyAttractionsRequest) ([]model.NearbyAttractionResponse, error) {
	if err := u.Validate.Struct(req); err != nil {
		return nil, fiber.ErrBadRequest
	}

	db := u.DB.WithContext(ctx)

	var property entity.Property
	if err := db.Select("id, lat, lng").Where("id = ?", propertyID).First(&property).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fiber.ErrNotFound
		}
		return nil, fiber.ErrInternalServerError
	}

	attractionIDs := make([]string, len(req.Attractions))
	for i, item := range req.Attractions {
		attractionIDs[i] = item.TouristAttractionID
	}

	var attractions []entity.TouristAttraction
	if err := db.Where("id IN ?", attractionIDs).Find(&attractions).Error; err != nil {
		return nil, fiber.ErrInternalServerError
	}
	attractionMap := make(map[string]*entity.TouristAttraction, len(attractions))
	for i := range attractions {
		attractionMap[attractions[i].ID] = &attractions[i]
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
			SortOrder:           item.SortOrder,
		}

		if property.Lat != nil && property.Lng != nil {
			if a, ok := attractionMap[item.TouristAttractionID]; ok && a.Latitude != nil && a.Longitude != nil {
				dist := haversineKm(*a.Latitude, *a.Longitude, *property.Lat, *property.Lng)
				distRounded := math.Round(dist*100) / 100
				row.DistanceKm = &distRounded
				// estimasi durasi berkendara ~40 km/jam
				mins := int(math.Round(dist / 40 * 60))
				row.DurationMinutes = &mins
			}
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
