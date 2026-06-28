package model

import (
	"id.diengs.backend/internal/entity"
)

type PropertyImageResponse struct {
	ID         string `json:"id,omitempty"`
	PropertyID string `json:"property_id,omitempty"`
	ImageURL   string `json:"image_url,omitempty"`
	IsPrimary  bool   `json:"is_primary,omitempty"`
	CreatedAt  int64  `json:"created_at,omitempty"`
	UpdatedAt  int64  `json:"updated_at,omitempty"`
}

type PropertyResponse struct {
	ID                string                     `json:"id"`
	Slug              string                     `json:"slug"`
	PropertyType      string                     `json:"property_type"`
	BookingType       string                     `json:"booking_type"`
	Title             string                     `json:"title"`
	Address           string                     `json:"address"`
	Description       string                     `json:"description"`
	ThumbnailURL      *string                    `json:"thumbnail_url,omitempty"`
	Lat               *float64                   `json:"lat,omitempty"`
	Lng               *float64                   `json:"lng,omitempty"`
	Images            []PropertyImageResponse    `json:"images,omitempty"`
	Host              HostProfileResponse        `json:"host,omitempty"`
	Rentable          []RentableResponse         `json:"rentable,omitempty"`
	Amenities         []AmenityResponse          `json:"amenities,omitempty"`
	NearbyAttractions []NearbyAttractionResponse `json:"nearby_attractions,omitempty"`
	CreatedAt         int64                      `json:"created_at"`
	UpdatedAt         int64                      `json:"updated_at"`
}

type SearchPropertyRequest struct {
	Key          string `query:"key"           validate:"max=100"`
	CheckIn      string `query:"check_in"`
	CheckOut     string `query:"check_out"`
	GuestCount   int    `query:"guest_count"`
	AttractionID string `query:"attraction_id"`
	PropertyType string `query:"property_type"`
	Page         int    `query:"page"          validate:"min=1"`
	Size         int    `query:"size"          validate:"min=1"`
}

type PropertyImageCreateRequest struct {
	ImageURL  string `json:"image_url" validate:"required"`
	IsPrimary bool   `json:"is_primary"`
}

type PropertyCreateRequest struct {
	HostID       *string                      `json:"host_id,omitempty"`
	Host         *HostCreateRequest           `json:"host,omitempty"`
	PropertyType string                       `json:"property_type"`
	BookingType  string                       `json:"booking_type"`
	Title        string                       `json:"title" validate:"required"`
	Slug         string                       `json:"slug,omitempty"`
	Address      string                       `json:"address" validate:"required"`
	Description  string                       `json:"description" validate:"required"`
	ThumbnailURL *string                      `json:"thumbnail_url,omitempty"`
	Lat          *float64                     `json:"lat,omitempty"`
	Lng          *float64                     `json:"lng,omitempty"`
	Images       []PropertyImageCreateRequest `json:"images,omitempty"`
	AmenityIDs   []string                     `json:"amenity_ids,omitempty"`
}

func PropertyToResponse(property *entity.Property) *PropertyResponse {
	host := HostToResponse(&property.Host)
	if host == nil {
		host = &HostProfileResponse{}
	}

	images := make([]PropertyImageResponse, 0, len(property.Images))
	for _, img := range property.Images {
		images = append(images, PropertyImageResponse{
			ID:         img.ID,
			PropertyID: img.PropertyID,
			ImageURL:   img.ImageURL,
			IsPrimary:  img.IsPrimary,
			CreatedAt:  img.CreatedAt,
			UpdatedAt:  img.UpdatedAt,
		})
	}

	return &PropertyResponse{
		ID:                property.ID,
		Slug:              property.Slug,
		Host:              *host,
		PropertyType:      property.PropertyType,
		BookingType:       property.BookingType,
		Title:             property.Title,
		Address:           property.Address,
		Description:       property.Description,
		ThumbnailURL:      property.ThumbnailURL,
		Lat:               property.Lat,
		Lng:               property.Lng,
		Images:            images,
		Rentable:          RentableToResponses(property.Rentable),
		Amenities:         AmenitiesToResponse(property.Amenities),
		NearbyAttractions: NearbyAttractionsToResponse(property.NearbyAttractions),
		CreatedAt:         property.CreatedAt,
		UpdatedAt:         property.UpdatedAt,
	}
}
