package repository

import (
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"id.diengs.backend/internal/entity"
)

type PropertyRepo struct {
	Repository[entity.Property]
	Log *logrus.Logger
}

func NewPropertyRepo(log *logrus.Logger) *PropertyRepo {
	return &PropertyRepo{
		Log: log,
	}
}

func (r *PropertyRepo) FindByHostID(db *gorm.DB, properties *[]entity.Property, hostID string) error {
	return db.Where("host_id = ?", hostID).Find(properties).Error
}

func (r *Repository[T]) FindByExperienceID(db *gorm.DB, entity *entity.Property, id any, preloads ...string) error {
	query := db

	for _, preload := range preloads {
		query = query.Preload(preload)
	}

	return query.Where("experience_id = ?", id).Take(entity).Error
}
