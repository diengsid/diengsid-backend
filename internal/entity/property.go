package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Property struct {
	ID           string  `gorm:"column:id;primaryKey"`
	HostID       string  `gorm:"column:host_id;not null"`
	PropertyType string  `gorm:"column:property_type;default:homestay"`
	BookingType  string  `gorm:"column:booking_type"`
	Title        string  `gorm:"column:title;not null"`
	Address      string  `gorm:"column:address;not null"`
	Description  string  `gorm:"column:description;type:text;not null"`
	ThumbnailURL *string `gorm:"column:thumbnail_url"`
	Lat          *float64 `gorm:"column:lat"`
	Lng          *float64 `gorm:"column:lng"`

	Host               HostProfile               `gorm:"foreignKey:HostID;references:ID;constraint:OnDelete:CASCADE"`
	Images             []PropertyImage           `gorm:"foreignKey:PropertyID;constraint:OnDelete:CASCADE"`
	Rentable           []Rentable
	Amenities          []Amenity                 `gorm:"many2many:property_amenities;"`
	NearbyAttractions  []PropertyNearbyAttraction `gorm:"foreignKey:PropertyID;constraint:OnDelete:CASCADE"`

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

type PropertyImage struct {
	ID         string `gorm:"column:id;primaryKey"`
	PropertyID string `gorm:"column:property_id;not null"`
	ImageURL   string `gorm:"column:image_url;not null"`
	IsPrimary  bool   `gorm:"column:is_primary;default:false"`

	Property Property `gorm:"foreignKey:PropertyID;references:ID;constraint:OnDelete:CASCADE"`

	CreatedAt int64 `gorm:"column:created_at"`
	UpdatedAt int64 `gorm:"column:updated_at"`
}

func (PropertyImage) TableName() string {
	return "property_images"
}

func (p *PropertyImage) BeforeCreate(tx *gorm.DB) (err error) {
	p.ID = uuid.NewString()
	p.CreatedAt = time.Now().UnixMilli()
	p.UpdatedAt = time.Now().UnixMilli()
	return nil
}
