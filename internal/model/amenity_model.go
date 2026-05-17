package model

import "id.diengs.backend/internal/entity"

type AmenityResponse struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Icon     string `json:"icon,omitempty"`
	Category string `json:"category,omitempty"`
}

type AmenityCreateRequest struct {
	Name     string `json:"name" validate:"required"`
	Icon     string `json:"icon"`
	Category string `json:"category"`
}

type SetAmenitiesRequest struct {
	AmenityIDs []string `json:"amenity_ids" validate:"required"`
}

func AmenityToResponse(a *entity.Amenity) AmenityResponse {
	return AmenityResponse{
		ID:       a.ID,
		Name:     a.Name,
		Icon:     a.Icon,
		Category: a.Category,
	}
}

func AmenitiesToResponse(amenities []entity.Amenity) []AmenityResponse {
	out := make([]AmenityResponse, len(amenities))
	for i := range amenities {
		out[i] = AmenityToResponse(&amenities[i])
	}
	return out
}
