package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Property struct {
	ID           string `gorm:"column:id;primaryKey"`
	HostID       string `gorm:"column:host_id;not null"`
	ExperienceID string `gorm:"column:experience_id;not null"`
	PropertyType string `gorm:"column:property_type;default:homestay"`
	BookingType  string `gorm:"column:booking_type"`

	Host       HostProfile `gorm:"foreignKey:HostID;references:ID;constraint:OnDelete:CASCADE"`
	Experience Experience  `gorm:"foreignKey:ExperienceID;references:ID;constraint:OnDelete:CASCADE"`
	Rentable   []Rentable

	CreatedAt int64 `gorm:"column:created_at"`
	UpdatedAt int64 `gorm:"column:updated_at"`
}

func (Property) TableName() string {
	return "properties"
}

func (p *Property) BeforeCreate(tx *gorm.DB) (err error) {
	p.ID = uuid.NewString()
	p.CreatedAt = time.Now().UnixMilli()
	p.UpdatedAt = time.Now().UnixMilli()
	return nil
}
