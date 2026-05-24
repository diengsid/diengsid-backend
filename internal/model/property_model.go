package model

import (
	"id.diengs.backend/internal/entity"
)

type PropertyResponse struct {
	ID           string              `json:"id"`
	PropertyType string              `json:"property_type"`
	BookingType  string              `json:"booking_type"`
	CreatedAt    int64               `json:"created_at"`
	UpdatedAt    int64               `json:"updated_at"`
	Experience   ExperienceResponse  `json:"experience,omitempty"`
	Host         HostProfileResponse `json:"host,omitempty"`
	Rentable     []RentableResponse  `json:"rentable,omitempty"`
	Amenities    []AmenityResponse   `json:"amenities,omitempty"`
}

type PropertyCreateRequest struct {
	ExperienceID string             `json:"experience_id" validate:"required"`
	HostID       *string            `json:"host_id,omitempty"`
	Host         *HostCreateRequest `json:"host,omitempty"`
	PropertyType string             `json:"property_type"`
	BookingType  string             `json:"booking_type"`
	AmenityIDs   []string           `json:"amenity_ids,omitempty"`
}

func PropertyToResponse(property *entity.Property) *PropertyResponse {
	experience := ExperienceToResponse(&property.Experience)
	if experience == nil {
		experience = &ExperienceResponse{}
	}

	host := HostToResponse(&property.Host)
	if host == nil {
		host = &HostProfileResponse{}
	}

	rentable := RentableToResponses(property.Rentable)

	return &PropertyResponse{
		ID:           property.ID,
		Experience:   *experience,
		Host:         *host,
		PropertyType: property.PropertyType,
		BookingType:  property.BookingType,
		CreatedAt:    property.CreatedAt,
		UpdatedAt:    property.UpdatedAt,
		Rentable:     rentable,
		Amenities:    AmenitiesToResponse(property.Amenities),
	}
}
