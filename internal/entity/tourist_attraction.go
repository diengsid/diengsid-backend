package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TouristAttraction struct {
	ID          string  `gorm:"column:id;primaryKey"`
	Name        string  `gorm:"column:name;not null"`
	Slug        string  `gorm:"column:slug;not null;unique"`
	Description *string `gorm:"column:description;type:text"`
	Address     *string `gorm:"column:address"`
	Latitude    *float64 `gorm:"column:latitude"`
	Longitude   *float64 `gorm:"column:longitude"`
	Category    *string `gorm:"column:category"`
	ImageURL    *string `gorm:"column:image_url"`

	CreatedAt int64 `gorm:"column:created_at"`
	UpdatedAt int64 `gorm:"column:updated_at"`
}

func (TouristAttraction) TableName() string {
	return "tourist_attractions"
}

func (t *TouristAttraction) BeforeCreate(tx *gorm.DB) error {
	t.ID = uuid.NewString()
	t.CreatedAt = time.Now().UnixMilli()
	t.UpdatedAt = time.Now().UnixMilli()
	return nil
}

// PropertyNearbyAttraction adalah join table dengan extra fields.
type PropertyNearbyAttraction struct {
	PropertyID           string  `gorm:"column:property_id;primaryKey"`
	TouristAttractionID  string  `gorm:"column:tourist_attraction_id;primaryKey"`
	DistanceKm           *float64 `gorm:"column:distance_km"`
	DurationMinutes      *int    `gorm:"column:duration_minutes"`
	SortOrder            int     `gorm:"column:sort_order;default:0"`

	TouristAttraction TouristAttraction `gorm:"foreignKey:TouristAttractionID;references:ID"`
}

func (PropertyNearbyAttraction) TableName() string {
	return "property_nearby_attractions"
}
