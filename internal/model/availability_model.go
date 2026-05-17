package model

type CheckAvailabilityRequest struct {
	CheckIn  string `query:"check_in" validate:"required"`
	CheckOut string `query:"check_out" validate:"required"`
}

type AvailabilityResponse struct {
	Date           string   `json:"date"`
	AvailableCount int      `json:"available_count"`
	PriceOverride  *float64 `json:"price_override,omitempty"`
	IsAvailable    bool     `json:"is_available"`
}
