package model

import "id.diengs.backend/internal/entity"

type ConfirmBookingRequest struct {
	// "WAITING_PAYMENT" or "UNAVAILABLE"
	Status string `json:"status" validate:"required,oneof=WAITING_PAYMENT UNAVAILABLE"`
}

type BookingCreateRequest struct {
	RentableID   string `json:"rentable_id" validate:"required"`
	PropertyID   string `json:"property_id" validate:"required"`
	CheckIn      string `json:"check_in" validate:"required"`  // format: 2006-01-02
	CheckOut     string `json:"check_out" validate:"required"` // format: 2006-01-02
	Quantity     int    `json:"quantity"`
	GuestCount   int    `json:"guest_count"`
	FirstPayment string `json:"first_payment"` // DP or FULL
}

type BookingResponse struct {
	ID            string  `json:"id"`
	UserID        string  `json:"user_id"`
	PropertyID    string  `json:"property_id"`
	RentableID    string  `json:"rentable_id"`
	Quantity      int     `json:"quantity"`
	GuestCount    int     `json:"guest_count"`
	CheckIn       string  `json:"check_in"`
	CheckOut      string  `json:"check_out"`
	TotalNight    int     `json:"total_night"`
	TotalPrice    float64 `json:"total_price"`
	Discount      float64 `json:"discount"`
	Status        string  `json:"status"`
	PaymentStatus string  `json:"payment_status"`
	FirstPayment  *string `json:"first_payment,omitempty"`
	CreatedAt     int64   `json:"created_at"`
	UpdatedAt     int64   `json:"updated_at"`
}

func BookingToResponse(b *entity.Booking) *BookingResponse {
	if b == nil {
		return nil
	}

	var firstPayment *string
	if b.FirstPayment != nil {
		v := string(*b.FirstPayment)
		firstPayment = &v
	}

	return &BookingResponse{
		ID:            b.ID,
		UserID:        b.UserID,
		PropertyID:    b.PropertyID,
		RentableID:    b.RentableID,
		Quantity:      b.Quantity,
		GuestCount:    b.GuestCount,
		CheckIn:       b.CheckIn.Format("2006-01-02"),
		CheckOut:      b.CheckOut.Format("2006-01-02"),
		TotalNight:    b.TotalNight,
		TotalPrice:    b.TotalPrice,
		Discount:      b.Discount,
		Status:        string(b.Status),
		PaymentStatus: string(b.PaymentStatus),
		FirstPayment:  firstPayment,
		CreatedAt:     b.CreatedAt,
		UpdatedAt:     b.UpdatedAt,
	}
}

func BookingsToResponses(bookings []entity.Booking) []BookingResponse {
	responses := make([]BookingResponse, len(bookings))
	for i := range bookings {
		responses[i] = *BookingToResponse(&bookings[i])
	}
	return responses
}
