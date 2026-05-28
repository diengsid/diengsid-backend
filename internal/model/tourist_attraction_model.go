package model

import "id.diengs.backend/internal/entity"

// ─── Request ──────────────────────────────────────────────────────────────────

type TouristAttractionCreateRequest struct {
	Name        string   `json:"name" validate:"required"`
	Slug        string   `json:"slug" validate:"required"`
	Description *string  `json:"description,omitempty"`
	Address     *string  `json:"address,omitempty"`
	Latitude    *float64 `json:"latitude,omitempty"`
	Longitude   *float64 `json:"longitude,omitempty"`
	Category    *string  `json:"category,omitempty"`
	ImageURL    *string  `json:"image_url,omitempty"`
}

type NearbyAttractionItem struct {
	TouristAttractionID string   `json:"tourist_attraction_id" validate:"required"`
	DistanceKm          *float64 `json:"distance_km,omitempty"`
	DurationMinutes     *int     `json:"duration_minutes,omitempty"`
	SortOrder           int      `json:"sort_order"`
}

type SetNearbyAttractionsRequest struct {
	Attractions []NearbyAttractionItem `json:"attractions" validate:"required"`
}

// ─── Response ─────────────────────────────────────────────────────────────────

type TouristAttractionResponse struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Slug        string   `json:"slug"`
	Description *string  `json:"description,omitempty"`
	Address     *string  `json:"address,omitempty"`
	Latitude    *float64 `json:"latitude,omitempty"`
	Longitude   *float64 `json:"longitude,omitempty"`
	Category    *string  `json:"category,omitempty"`
	ImageURL    *string  `json:"image_url,omitempty"`
	CreatedAt   int64    `json:"created_at"`
	UpdatedAt   int64    `json:"updated_at"`
}

type NearbyAttractionResponse struct {
	TouristAttractionID string   `json:"tourist_attraction_id"`
	DistanceKm          *float64 `json:"distance_km,omitempty"`
	DurationMinutes     *int     `json:"duration_minutes,omitempty"`
	SortOrder           int      `json:"sort_order"`
	Attraction          TouristAttractionResponse `json:"attraction"`
}

func TouristAttractionToResponse(a *entity.TouristAttraction) TouristAttractionResponse {
	return TouristAttractionResponse{
		ID:          a.ID,
		Name:        a.Name,
		Slug:        a.Slug,
		Description: a.Description,
		Address:     a.Address,
		Latitude:    a.Latitude,
		Longitude:   a.Longitude,
		Category:    a.Category,
		ImageURL:    a.ImageURL,
		CreatedAt:   a.CreatedAt,
		UpdatedAt:   a.UpdatedAt,
	}
}

func NearbyAttractionToResponse(n *entity.PropertyNearbyAttraction) NearbyAttractionResponse {
	return NearbyAttractionResponse{
		TouristAttractionID: n.TouristAttractionID,
		DistanceKm:          n.DistanceKm,
		DurationMinutes:     n.DurationMinutes,
		SortOrder:           n.SortOrder,
		Attraction:          TouristAttractionToResponse(&n.TouristAttraction),
	}
}

func NearbyAttractionsToResponse(list []entity.PropertyNearbyAttraction) []NearbyAttractionResponse {
	out := make([]NearbyAttractionResponse, len(list))
	for i := range list {
		out[i] = NearbyAttractionToResponse(&list[i])
	}
	return out
}
