package repository

import (
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"id.diengs.backend/internal/entity"
)

type TouristAttractionRepo struct {
	Repository[entity.TouristAttraction]
	Log *logrus.Logger
}

func NewTouristAttractionRepo(log *logrus.Logger) *TouristAttractionRepo {
	return &TouristAttractionRepo{Log: log}
}

func (r *TouristAttractionRepo) FindAll(db *gorm.DB, list *[]entity.TouristAttraction) error {
	return db.Order("category, name").Find(list).Error
}

func (r *TouristAttractionRepo) FindBySlug(db *gorm.DB, attraction *entity.TouristAttraction, slug string) error {
	return db.Where("slug = ?", slug).Take(attraction).Error
}

func (r *TouristAttractionRepo) FindNearbyByPropertyID(db *gorm.DB, list *[]entity.PropertyNearbyAttraction, propertyID string) error {
	return db.
		Preload("TouristAttraction").
		Where("property_id = ?", propertyID).
		Order("sort_order ASC").
		Find(list).Error
}
