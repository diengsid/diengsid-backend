package repository

import (
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"id.diengs.backend/internal/entity"
)

type BookingRepo struct {
	Repository[entity.Booking]
	Log *logrus.Logger
}

func NewBookingRepo(log *logrus.Logger) *BookingRepo {
	return &BookingRepo{Log: log}
}

func (r *BookingRepo) FindByUserID(db *gorm.DB, bookings *[]entity.Booking, userID string) error {
	return db.Where("user_id = ?", userID).Order("created_at DESC").Find(bookings).Error
}

func (r *BookingRepo) FindByPropertyIDs(db *gorm.DB, bookings *[]entity.Booking, propertyIDs []string) error {
	return db.Where("property_id IN ?", propertyIDs).Order("created_at DESC").Find(bookings).Error
}

func (r *BookingRepo) FindAll(db *gorm.DB, bookings *[]entity.Booking) error {
	return db.Order("created_at DESC").Find(bookings).Error
}
