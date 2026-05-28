package repository

import (
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"id.diengs.backend/internal/entity"
	"id.diengs.backend/internal/model"
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

func (r *PropertyRepo) Search(
	db *gorm.DB,
	req *model.SearchPropertyRequest,
	checkInUnix, checkOutUnix int64,
	guestCount int,
) ([]entity.Property, int64, error) {
	base := r.buildSearchQuery(db, req, checkInUnix, checkOutUnix, guestCount)

	var properties []entity.Property
	if err := base.
		Preload("Images").
		Preload("Host").
		Preload("Amenities").
		Preload("NearbyAttractions").
		Preload("NearbyAttractions.TouristAttraction").
		Order("properties.created_at DESC").
		Offset((req.Page - 1) * req.Size).
		Limit(req.Size).
		Find(&properties).Error; err != nil {
		return nil, 0, err
	}

	var total int64
	if err := r.buildSearchQuery(db, req, checkInUnix, checkOutUnix, guestCount).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	return properties, total, nil
}

func (r *PropertyRepo) buildSearchQuery(
	db *gorm.DB,
	req *model.SearchPropertyRequest,
	checkInUnix, checkOutUnix int64,
	guestCount int,
) *gorm.DB {
	query := db.Model(&entity.Property{})

	if req.Key != "" {
		key := "%" + req.Key + "%"
		query = query.Where("(properties.title ILIKE ? OR properties.address ILIKE ?)", key, key)
	}

	if req.PropertyType != "" {
		query = query.Where("properties.property_type = ?", req.PropertyType)
	}

	// Filter ketersediaan: property valid jika ada minimal satu rentable yang:
	// 1. capacity >= guestCount  → satu kamar bisa menampung semua tamu
	// 2. stock >= 1              → default ada kamar tersedia (fallback saat tidak ada record)
	// 3. NOT EXISTS tanggal di range dengan available_count < 1
	//    → tidak ada tanggal yang semua kamarnya sudah penuh (fully booked)
	if checkInUnix > 0 && checkOutUnix > 0 {
		query = query.Where(`EXISTS (
			SELECT 1 FROM rentables r
			WHERE r.property_id = properties.id
			  AND r.capacity >= ?
			  AND r.stock >= 1
			  AND NOT EXISTS (
				SELECT 1 FROM availabilities a
				WHERE a.rentable_id = r.id
				  AND a.date >= ?
				  AND a.date < ?
				  AND a.available_count < 1
			)
		)`, guestCount, checkInUnix, checkOutUnix)
	}

	// Filter objek wisata terdekat
	if req.AttractionID != "" {
		query = query.Where(`EXISTS (
			SELECT 1 FROM property_nearby_attractions pna
			WHERE pna.property_id = properties.id
			  AND pna.tourist_attraction_id = ?
		)`, req.AttractionID)
	}

	return query
}
