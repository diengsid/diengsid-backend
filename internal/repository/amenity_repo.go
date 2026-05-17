package repository

import (
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"id.diengs.backend/internal/entity"
)

type AmenityRepo struct {
	Repository[entity.Amenity]
	Log *logrus.Logger
}

func NewAmenityRepo(log *logrus.Logger) *AmenityRepo {
	return &AmenityRepo{Log: log}
}

func (r *AmenityRepo) FindAll(db *gorm.DB, amenities *[]entity.Amenity) error {
	return db.Order("category, name").Find(amenities).Error
}

func (r *AmenityRepo) FindByIDs(db *gorm.DB, amenities *[]entity.Amenity, ids []string) error {
	return db.Where("id IN ?", ids).Find(amenities).Error
}
