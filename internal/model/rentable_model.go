package model

import "id.diengs.backend/internal/entity"

type RentableCreateRequest struct {
	PropertyID string  `json:"property_id"`
	Type       string  `json:"type"`
	Name       string  `json:"name"`
	ImageUrl   string  `json:"image_url"`
	Capacity   int     `json:"capacity"`
	BasePrice  float64 `json:"base_price"`
	Discount   float64 `json:"discount"`
	Stock      int     `json:"stock"`
}

type RentableResponse struct {
	ID         string  `json:"id"`
	PropertyID string  `json:"property_id"`
	Type       string  `json:"type"`
	Name       string  `json:"name"`
	ImageUrl   string  `json:"image_url"`
	Capacity   int     `json:"capacity"`
	BasePrice  float64 `json:"base_price"`
	Discount   float64 `json:"discount"`
	Stock      int     `json:"stock"`
	CreatedAt  int64   `json:"created_at"`
	UpdatedAt  int64   `json:"updated_at"`
}

func RentableToResponse(rentable *entity.Rentable) *RentableResponse {
	if rentable == nil {
		return nil
	}

	return &RentableResponse{
		ID:         rentable.ID,
		PropertyID: rentable.PropertyID,
		Type:       rentable.Type,
		Name:       rentable.Name,
		ImageUrl:   rentable.ImageUrl,
		Capacity:   rentable.Capacity,
		BasePrice:  rentable.BasePrice,
		Discount:   rentable.Discount,
		Stock:      rentable.Stock,
		CreatedAt:  rentable.CreatedAt,
		UpdatedAt:  rentable.UpdatedAt,
	}
}

func RentableToResponses(rentables []entity.Rentable) []RentableResponse {
	responses := make([]RentableResponse, len(rentables))
	for i := range rentables {
		responses[i] = *RentableToResponse(&rentables[i])
	}

	return responses
}
