package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Amenity struct {
	ID        string `gorm:"column:id;primaryKey"`
	Name      string `gorm:"column:name;not null"`
	Icon      string `gorm:"column:icon"`
	Category  string `gorm:"column:category"`
	CreatedAt int64  `gorm:"column:created_at"`
	UpdatedAt int64  `gorm:"column:updated_at"`
}

func (Amenity) TableName() string {
	return "amenities"
}

func (a *Amenity) BeforeCreate(tx *gorm.DB) error {
	a.ID = uuid.NewString()
	a.CreatedAt = time.Now().UnixMilli()
	a.UpdatedAt = time.Now().UnixMilli()
	return nil
}
