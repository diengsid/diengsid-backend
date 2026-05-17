package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Rentable struct {
	ID         string  `gorm:"column:id;primaryKey"`
	PropertyID string  `gorm:"column:property_id;not null"`
	Type       string  `gorm:"column:type"`
	Name       string  `gorm:"column:name"`
	ImageUrl   string  `gorm:"image_url"`
	Capacity   int     `gorm:"column:capacity"`
	BasePrice  float64 `gorm:"column:base_price;not null"`
	Discount   float64 `gorm:"column:discount"`
	Stock      int     `gorm:"column:stock;not null;default:1"`
	Property   Property

	CreatedAt int64 `gorm:"column:created_at"`
	UpdatedAt int64 `gorm:"column:updated_at"`
}

func (Rentable) TableName() string {
	return "rentables"
}

func (r *Rentable) BeforeCreate(tx *gorm.DB) (err error) {
	r.ID = uuid.NewString()
	r.CreatedAt = time.Now().UnixMilli()
	r.UpdatedAt = time.Now().UnixMilli()
	return nil
}
