package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Availability struct {
	ID             string   `gorm:"column:id;primaryKey"`
	RentableID     string   `gorm:"column:rentable_id;not null"`
	Date           int64    `gorm:"column:date;not null"`
	AvailableCount int      `gorm:"column:available_count;not null"`
	PriceOverride  *float64 `gorm:"column:price_override"`

	Rentable Rentable `gorm:"foreignKey:RentableID;references:ID;constraint:OnDelete:CASCADE"`

	CreatedAt int64 `gorm:"column:created_at"`
	UpdatedAt int64 `gorm:"column:updated_at"`
}

func (Availability) TableName() string {
	return "availabilities"
}

func (a *Availability) BeforeCreate(tx *gorm.DB) error {
	a.ID = uuid.NewString()
	a.CreatedAt = time.Now().UnixMilli()
	a.UpdatedAt = time.Now().UnixMilli()
	return nil
}
