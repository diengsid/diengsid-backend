package repository

import (
	"time"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"id.diengs.backend/internal/entity"
)

type AvailabilityRepo struct {
	Repository[entity.Availability]
	Log *logrus.Logger
}

func NewAvailabilityRepo(log *logrus.Logger) *AvailabilityRepo {
	return &AvailabilityRepo{Log: log}
}

func (r *AvailabilityRepo) FindByRentableAndDateRange(db *gorm.DB, avails *[]entity.Availability, rentableID string, checkIn, checkOut time.Time) error {
	return db.Where("rentable_id = ? AND date >= ? AND date < ?", rentableID, checkIn.Unix(), checkOut.Unix()).
		Find(avails).Error
}

// Upsert menggunakan ON CONFLICT untuk atomic insert-or-update.
func (r *AvailabilityRepo) Upsert(db *gorm.DB, avail *entity.Availability) error {
	avail.UpdatedAt = time.Now().UnixMilli()
	return db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "rentable_id"}, {Name: "date"}},
		DoUpdates: clause.AssignmentColumns([]string{"available_count", "updated_at"}),
	}).Create(avail).Error
}
