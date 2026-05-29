package test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"id.diengs.backend/internal/entity"
	"id.diengs.backend/internal/model"
)

func doCreateBooking(t *testing.T, tokenCookie string, req model.BookingCreateRequest) *http.Response {
	t.Helper()
	body, _ := json.Marshal(req)
	r := httptest.NewRequest(http.MethodPost, "/api/bookings", bytes.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	if tokenCookie != "" {
		r.Header.Set("Cookie", "token="+tokenCookie)
	}
	resp, err := app.Test(r)
	assert.Nil(t, err)
	return resp
}

// ─── Create Booking ───────────────────────────────────────────────────────────

func TestCreateBooking_Success(t *testing.T) {
	ClearAll()
	rentable := SeedRentableWithDeps()
	token := RegisterAndGetCookie(t)

	resp := doCreateBooking(t, token, model.BookingCreateRequest{
		RentableID: rentable.ID,
		PropertyID: rentable.PropertyID,
		CheckIn:    "2026-08-01",
		CheckOut:   "2026-08-04",
		PhoneNumber: "08123456789",
		Quantity:   1,
	})

	body, _ := io.ReadAll(resp.Body)
	var respBody model.WebResponse[*model.BookingResponse]
	json.Unmarshal(body, &respBody)

	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	assert.True(t, respBody.Success)
	assert.NotEmpty(t, respBody.Data.ID)
	assert.Equal(t, 3, respBody.Data.TotalNight)
	// (200000 - 10%) * 3 malam * 1 = 540.000
	assert.Equal(t, 540_000.0, respBody.Data.TotalPrice)
	assert.Equal(t, "PENDING", respBody.Data.Status)
	assert.Equal(t, "UNPAID", respBody.Data.PaymentStatus)
}

func TestCreateBooking_Unauthorized(t *testing.T) {
	ClearAll()
	rentable := SeedRentableWithDeps()

	resp := doCreateBooking(t, "", model.BookingCreateRequest{
		RentableID: rentable.ID,
		PropertyID: rentable.PropertyID,
		CheckIn:    "2026-08-01",
		CheckOut:   "2026-08-04",
		PhoneNumber: "08123456789",
	})

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestCreateBooking_ValidationError(t *testing.T) {
	ClearAll()
	token := RegisterAndGetCookie(t)

	resp := doCreateBooking(t, token, model.BookingCreateRequest{
		PropertyID: "some-property-id",
		CheckIn:    "2026-08-01",
		CheckOut:   "2026-08-04",
		PhoneNumber: "08123456789",
	})

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestCreateBooking_InvalidDateFormat(t *testing.T) {
	ClearAll()
	rentable := SeedRentableWithDeps()
	token := RegisterAndGetCookie(t)

	resp := doCreateBooking(t, token, model.BookingCreateRequest{
		RentableID: rentable.ID,
		PropertyID: rentable.PropertyID,
		CheckIn:    "01-08-2026",
		CheckOut:   "04-08-2026",
		PhoneNumber: "08123456789",
	})

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestCreateBooking_CheckoutBeforeCheckin(t *testing.T) {
	ClearAll()
	rentable := SeedRentableWithDeps()
	token := RegisterAndGetCookie(t)

	resp := doCreateBooking(t, token, model.BookingCreateRequest{
		RentableID: rentable.ID,
		PropertyID: rentable.PropertyID,
		CheckIn:    "2026-08-10",
		CheckOut:   "2026-08-05",
		PhoneNumber: "08123456789",
	})

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

// ─── Get Booking by ID ────────────────────────────────────────────────────────

func TestGetBookingByID_Success(t *testing.T) {
	ClearAll()
	rentable := SeedRentableWithDeps()
	token := RegisterAndGetCookie(t)

	createResp := doCreateBooking(t, token, model.BookingCreateRequest{
		RentableID: rentable.ID,
		PropertyID: rentable.PropertyID,
		CheckIn:    "2026-09-01",
		CheckOut:   "2026-09-03",
		PhoneNumber: "08123456789",
		Quantity:   1,
	})
	assert.Equal(t, http.StatusCreated, createResp.StatusCode)

	var createBody model.WebResponse[*model.BookingResponse]
	data, _ := io.ReadAll(createResp.Body)
	json.Unmarshal(data, &createBody)
	bookingID := createBody.Data.ID

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/bookings/%s", bookingID), nil)
	req.Header.Set("Cookie", "token="+token)
	resp, err := app.Test(req)

	assert.Nil(t, err)
	body, _ := io.ReadAll(resp.Body)
	var respBody model.WebResponse[*model.BookingResponse]
	json.Unmarshal(body, &respBody)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, bookingID, respBody.Data.ID)
	assert.Equal(t, 2, respBody.Data.TotalNight)
}

func TestGetBookingByID_NotFound(t *testing.T) {
	ClearAll()
	token := RegisterAndGetCookie(t)

	req := httptest.NewRequest(http.MethodGet, "/api/bookings/booking-yang-tidak-ada", nil)
	req.Header.Set("Cookie", "token="+token)
	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestGetBookingByID_Forbidden(t *testing.T) {
	ClearAll()
	rentable := SeedRentableWithDeps()

	// user 1 buat booking
	token1 := RegisterAndGetCookie(t)
	createResp := doCreateBooking(t, token1, model.BookingCreateRequest{
		RentableID: rentable.ID,
		PropertyID: rentable.PropertyID,
		CheckIn:    "2026-09-01",
		CheckOut:   "2026-09-03",
		PhoneNumber: "08123456789",
	})
	var createBody model.WebResponse[*model.BookingResponse]
	data, _ := io.ReadAll(createResp.Body)
	json.Unmarshal(data, &createBody)
	bookingID := createBody.Data.ID

	// user 2 register & ambil cookie
	regResp := DoRegister(t, "User Dua", "user2@example.com", "password123")
	token2 := cookieFromResp(regResp, "token")

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/bookings/%s", bookingID), nil)
	req.Header.Set("Cookie", "token="+token2)
	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
}

// ─── Get My Bookings ──────────────────────────────────────────────────────────

func TestGetMyBookings_Success(t *testing.T) {
	ClearAll()
	rentable := SeedRentableWithDeps()
	token := RegisterAndGetCookie(t)

	doCreateBooking(t, token, model.BookingCreateRequest{
		RentableID: rentable.ID,
		PropertyID: rentable.PropertyID,
		CheckIn:    "2026-10-01",
		CheckOut:   "2026-10-03",
		PhoneNumber: "08123456789",
	})
	doCreateBooking(t, token, model.BookingCreateRequest{
		RentableID: rentable.ID,
		PropertyID: rentable.PropertyID,
		CheckIn:    "2026-11-01",
		CheckOut:   "2026-11-04",
		PhoneNumber: "08123456789",
	})

	req := httptest.NewRequest(http.MethodGet, "/api/bookings/my", nil)
	req.Header.Set("Cookie", "token="+token)
	resp, err := app.Test(req)

	assert.Nil(t, err)
	body, _ := io.ReadAll(resp.Body)
	var respBody model.WebResponse[[]model.BookingResponse]
	json.Unmarshal(body, &respBody)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Len(t, respBody.Data, 2)
}

func TestGetMyBookings_Unauthorized(t *testing.T) {
	ClearAll()

	req := httptest.NewRequest(http.MethodGet, "/api/bookings/my", nil)
	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

// ─── Mark Done (REVIEW → DONE) ───────────────────────────────────────────────

func TestComplete_Success(t *testing.T) {
	ClearAll()
	rentable := SeedRentableWithDeps()
	token := RegisterAndGetCookie(t)

	createResp := doCreateBooking(t, token, model.BookingCreateRequest{
		RentableID: rentable.ID,
		PropertyID: rentable.PropertyID,
		CheckIn:    "2026-09-01",
		CheckOut:   "2026-09-03",
		PhoneNumber: "08123456789",
		Quantity:   1,
	})
	var createBody model.WebResponse[*model.BookingResponse]
	data, _ := io.ReadAll(createResp.Body)
	json.Unmarshal(data, &createBody)
	bookingID := createBody.Data.ID

	// Force status ke REVIEW langsung di DB
	db.Model(&entity.Booking{}).Where("id = ?", bookingID).Update("status", entity.StatusReview)

	req := httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/api/bookings/%s/done", bookingID), nil)
	req.Header.Set("Cookie", "token="+token)
	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var respBody model.WebResponse[*model.BookingResponse]
	body, _ := io.ReadAll(resp.Body)
	json.Unmarshal(body, &respBody)
	assert.Equal(t, "DONE", respBody.Data.Status)
}

func TestComplete_WrongStatus(t *testing.T) {
	ClearAll()
	rentable := SeedRentableWithDeps()
	token := RegisterAndGetCookie(t)

	createResp := doCreateBooking(t, token, model.BookingCreateRequest{
		RentableID: rentable.ID,
		PropertyID: rentable.PropertyID,
		CheckIn:    "2026-09-01",
		CheckOut:   "2026-09-03",
		PhoneNumber: "08123456789",
	})
	var createBody model.WebResponse[*model.BookingResponse]
	data, _ := io.ReadAll(createResp.Body)
	json.Unmarshal(data, &createBody)
	bookingID := createBody.Data.ID

	// Status masih PENDING, bukan REVIEW
	req := httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/api/bookings/%s/done", bookingID), nil)
	req.Header.Set("Cookie", "token="+token)
	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestComplete_Forbidden(t *testing.T) {
	ClearAll()
	rentable := SeedRentableWithDeps()
	token1 := RegisterAndGetCookie(t)

	createResp := doCreateBooking(t, token1, model.BookingCreateRequest{
		RentableID: rentable.ID,
		PropertyID: rentable.PropertyID,
		CheckIn:    "2026-09-01",
		CheckOut:   "2026-09-03",
		PhoneNumber: "08123456789",
	})
	var createBody model.WebResponse[*model.BookingResponse]
	data, _ := io.ReadAll(createResp.Body)
	json.Unmarshal(data, &createBody)
	bookingID := createBody.Data.ID

	db.Model(&entity.Booking{}).Where("id = ?", bookingID).Update("status", entity.StatusReview)

	// User lain coba selesaikan booking orang lain
	token2 := RegisterWithEmail(t, "User Dua", "user2@example.com")
	req := httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/api/bookings/%s/done", bookingID), nil)
	req.Header.Set("Cookie", "token="+token2)
	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
}

// ─── Host: Confirm Booking ────────────────────────────────────────────────────

func TestConfirmBooking_Success_WaitingPayment(t *testing.T) {
	ClearAll()
	rentable := SeedRentableWithDeps()

	// Daftarkan user tamu
	guestToken := RegisterWithEmail(t, "Tamu Satu", "tamu@example.com")
	createResp := doCreateBooking(t, guestToken, model.BookingCreateRequest{
		RentableID: rentable.ID,
		PropertyID: rentable.PropertyID,
		CheckIn:    "2026-09-10",
		CheckOut:   "2026-09-12",
		PhoneNumber: "08123456789",
	})
	var createBody model.WebResponse[*model.BookingResponse]
	data, _ := io.ReadAll(createResp.Body)
	json.Unmarshal(data, &createBody)
	bookingID := createBody.Data.ID

	// Daftarkan host dengan email yang sama dengan host_profiles
	hostToken := RegisterWithEmail(t, "Host Test", "host@test.com")

	body, _ := json.Marshal(model.ConfirmBookingRequest{Status: "WAITING_PAYMENT"})
	req := httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/api/host/bookings/%s/confirm", bookingID), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cookie", "token="+hostToken)
	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var respBody model.WebResponse[*model.BookingResponse]
	respData, _ := io.ReadAll(resp.Body)
	json.Unmarshal(respData, &respBody)
	assert.Equal(t, "WAITING_PAYMENT", respBody.Data.Status)
}

func TestConfirmBooking_Success_Unavailable(t *testing.T) {
	ClearAll()
	rentable := SeedRentableWithDeps()

	guestToken := RegisterWithEmail(t, "Tamu Dua", "tamu2@example.com")
	createResp := doCreateBooking(t, guestToken, model.BookingCreateRequest{
		RentableID: rentable.ID,
		PropertyID: rentable.PropertyID,
		CheckIn:    "2026-10-01",
		CheckOut:   "2026-10-03",
		PhoneNumber: "08123456789",
	})
	var createBody model.WebResponse[*model.BookingResponse]
	data, _ := io.ReadAll(createResp.Body)
	json.Unmarshal(data, &createBody)
	bookingID := createBody.Data.ID

	hostToken := RegisterWithEmail(t, "Host Test", "host@test.com")

	body, _ := json.Marshal(model.ConfirmBookingRequest{Status: "UNAVAILABLE"})
	req := httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/api/host/bookings/%s/confirm", bookingID), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cookie", "token="+hostToken)
	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var respBody model.WebResponse[*model.BookingResponse]
	respData, _ := io.ReadAll(resp.Body)
	json.Unmarshal(respData, &respBody)
	assert.Equal(t, "UNAVAILABLE", respBody.Data.Status)
}

func TestConfirmBooking_Forbidden_NotOwner(t *testing.T) {
	ClearAll()
	rentable := SeedRentableWithDeps()

	guestToken := RegisterWithEmail(t, "Tamu Tiga", "tamu3@example.com")
	createResp := doCreateBooking(t, guestToken, model.BookingCreateRequest{
		RentableID: rentable.ID,
		PropertyID: rentable.PropertyID,
		CheckIn:    "2026-10-05",
		CheckOut:   "2026-10-07",
		PhoneNumber: "08123456789",
	})
	var createBody model.WebResponse[*model.BookingResponse]
	data, _ := io.ReadAll(createResp.Body)
	json.Unmarshal(data, &createBody)
	bookingID := createBody.Data.ID

	// User biasa (bukan host properti) coba konfirmasi
	body, _ := json.Marshal(model.ConfirmBookingRequest{Status: "WAITING_PAYMENT"})
	req := httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/api/host/bookings/%s/confirm", bookingID), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cookie", "token="+guestToken)
	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
}

func TestConfirmBooking_WrongStatus(t *testing.T) {
	ClearAll()
	rentable := SeedRentableWithDeps()

	guestToken := RegisterWithEmail(t, "Tamu Empat", "tamu4@example.com")
	createResp := doCreateBooking(t, guestToken, model.BookingCreateRequest{
		RentableID: rentable.ID,
		PropertyID: rentable.PropertyID,
		CheckIn:    "2026-10-10",
		CheckOut:   "2026-10-12",
		PhoneNumber: "08123456789",
	})
	var createBody model.WebResponse[*model.BookingResponse]
	data, _ := io.ReadAll(createResp.Body)
	json.Unmarshal(data, &createBody)
	bookingID := createBody.Data.ID

	// Ubah status ke WAITING_PAYMENT dulu
	db.Model(&entity.Booking{}).Where("id = ?", bookingID).Update("status", entity.StatusWaiting)

	hostToken := RegisterWithEmail(t, "Host Test", "host@test.com")

	body, _ := json.Marshal(model.ConfirmBookingRequest{Status: "WAITING_PAYMENT"})
	req := httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/api/host/bookings/%s/confirm", bookingID), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cookie", "token="+hostToken)
	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

// ─── Host: Checkout ───────────────────────────────────────────────────────────

func TestCheckout_Success(t *testing.T) {
	ClearAll()
	rentable := SeedRentableWithDeps()

	guestToken := RegisterWithEmail(t, "Tamu Lima", "tamu5@example.com")
	createResp := doCreateBooking(t, guestToken, model.BookingCreateRequest{
		RentableID: rentable.ID,
		PropertyID: rentable.PropertyID,
		CheckIn:    "2026-11-01",
		CheckOut:   "2026-11-04",
		PhoneNumber: "08123456789",
	})
	var createBody model.WebResponse[*model.BookingResponse]
	data, _ := io.ReadAll(createResp.Body)
	json.Unmarshal(data, &createBody)
	bookingID := createBody.Data.ID

	db.Model(&entity.Booking{}).Where("id = ?", bookingID).Update("status", entity.StatusCheckIn)

	hostToken := RegisterWithEmail(t, "Host Test", "host@test.com")

	req := httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/api/host/bookings/%s/checkout", bookingID), nil)
	req.Header.Set("Cookie", "token="+hostToken)
	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var respBody model.WebResponse[*model.BookingResponse]
	body, _ := io.ReadAll(resp.Body)
	json.Unmarshal(body, &respBody)
	assert.Equal(t, "REVIEW", respBody.Data.Status)
}

func TestCheckout_WrongStatus(t *testing.T) {
	ClearAll()
	rentable := SeedRentableWithDeps()

	guestToken := RegisterWithEmail(t, "Tamu Enam", "tamu6@example.com")
	createResp := doCreateBooking(t, guestToken, model.BookingCreateRequest{
		RentableID: rentable.ID,
		PropertyID: rentable.PropertyID,
		CheckIn:    "2026-11-10",
		CheckOut:   "2026-11-12",
		PhoneNumber: "08123456789",
	})
	var createBody model.WebResponse[*model.BookingResponse]
	data, _ := io.ReadAll(createResp.Body)
	json.Unmarshal(data, &createBody)
	bookingID := createBody.Data.ID

	// Status masih PENDING, bukan CHECK_IN
	hostToken := RegisterWithEmail(t, "Host Test", "host@test.com")

	req := httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/api/host/bookings/%s/checkout", bookingID), nil)
	req.Header.Set("Cookie", "token="+hostToken)
	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

// ─── Host: Get Host Bookings ──────────────────────────────────────────────────

func TestGetHostBookings_Success(t *testing.T) {
	ClearAll()
	rentable := SeedRentableWithDeps()

	guestToken := RegisterWithEmail(t, "Tamu Tujuh", "tamu7@example.com")
	doCreateBooking(t, guestToken, model.BookingCreateRequest{
		RentableID: rentable.ID,
		PropertyID: rentable.PropertyID,
		CheckIn:    "2026-12-01",
		CheckOut:   "2026-12-03",
		PhoneNumber: "08123456789",
	})

	hostToken := RegisterWithEmail(t, "Host Test", "host@test.com")

	req := httptest.NewRequest(http.MethodGet, "/api/host/bookings", nil)
	req.Header.Set("Cookie", "token="+hostToken)
	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var respBody model.WebResponse[[]model.BookingResponse]
	body, _ := io.ReadAll(resp.Body)
	json.Unmarshal(body, &respBody)
	assert.GreaterOrEqual(t, len(respBody.Data), 1)
}

// ─── Admin: Get All Bookings ──────────────────────────────────────────────────

func TestGetAllBookings_Admin_Success(t *testing.T) {
	ClearAll()
	rentable := SeedRentableWithDeps()
	guestToken := RegisterWithEmail(t, "Tamu Delapan", "tamu8@example.com")
	doCreateBooking(t, guestToken, model.BookingCreateRequest{
		RentableID: rentable.ID,
		PropertyID: rentable.PropertyID,
		CheckIn:    "2026-12-10",
		CheckOut:   "2026-12-12",
		PhoneNumber: "08123456789",
	})

	adminToken := SeedAdminAndGetToken(t)

	req := httptest.NewRequest(http.MethodGet, "/api/admin/bookings", nil)
	req.Header.Set("Cookie", "token="+adminToken)
	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var respBody model.WebResponse[[]model.BookingResponse]
	body, _ := io.ReadAll(resp.Body)
	json.Unmarshal(body, &respBody)
	assert.GreaterOrEqual(t, len(respBody.Data), 1)
}

func TestGetAllBookings_Admin_Forbidden(t *testing.T) {
	ClearAll()
	token := RegisterAndGetCookie(t)

	req := httptest.NewRequest(http.MethodGet, "/api/admin/bookings", nil)
	req.Header.Set("Cookie", "token="+token)
	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
}

// ─── Admin: Confirm & Checkout & Complete ────────────────────────────────────

func TestAdminConfirmBooking_Success(t *testing.T) {
	ClearAll()
	rentable := SeedRentableWithDeps()
	guestToken := RegisterWithEmail(t, "Tamu Sembilan", "tamu9@example.com")
	createResp := doCreateBooking(t, guestToken, model.BookingCreateRequest{
		RentableID: rentable.ID,
		PropertyID: rentable.PropertyID,
		CheckIn:    "2026-12-15",
		CheckOut:   "2026-12-17",
		PhoneNumber: "08123456789",
	})
	var createBody model.WebResponse[*model.BookingResponse]
	data, _ := io.ReadAll(createResp.Body)
	json.Unmarshal(data, &createBody)
	bookingID := createBody.Data.ID

	adminToken := SeedAdminAndGetToken(t)

	body, _ := json.Marshal(model.ConfirmBookingRequest{Status: "WAITING_PAYMENT"})
	req := httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/api/admin/bookings/%s/confirm", bookingID), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cookie", "token="+adminToken)
	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var respBody model.WebResponse[*model.BookingResponse]
	respData, _ := io.ReadAll(resp.Body)
	json.Unmarshal(respData, &respBody)
	assert.Equal(t, "WAITING_PAYMENT", respBody.Data.Status)
}

func TestAdminCheckout_Success(t *testing.T) {
	ClearAll()
	rentable := SeedRentableWithDeps()
	guestToken := RegisterWithEmail(t, "Tamu Sepuluh", "tamu10@example.com")
	createResp := doCreateBooking(t, guestToken, model.BookingCreateRequest{
		RentableID: rentable.ID,
		PropertyID: rentable.PropertyID,
		CheckIn:    "2026-12-20",
		CheckOut:   "2026-12-22",
		PhoneNumber: "08123456789",
	})
	var createBody model.WebResponse[*model.BookingResponse]
	data, _ := io.ReadAll(createResp.Body)
	json.Unmarshal(data, &createBody)
	bookingID := createBody.Data.ID

	db.Model(&entity.Booking{}).Where("id = ?", bookingID).Update("status", entity.StatusCheckIn)

	adminToken := SeedAdminAndGetToken(t)

	req := httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/api/admin/bookings/%s/checkout", bookingID), nil)
	req.Header.Set("Cookie", "token="+adminToken)
	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var respBody model.WebResponse[*model.BookingResponse]
	body, _ := io.ReadAll(resp.Body)
	json.Unmarshal(body, &respBody)
	assert.Equal(t, "REVIEW", respBody.Data.Status)
}

func TestAdminComplete_Success(t *testing.T) {
	ClearAll()
	rentable := SeedRentableWithDeps()
	guestToken := RegisterWithEmail(t, "Tamu Sebelas", "tamu11@example.com")
	createResp := doCreateBooking(t, guestToken, model.BookingCreateRequest{
		RentableID: rentable.ID,
		PropertyID: rentable.PropertyID,
		CheckIn:    "2026-12-25",
		CheckOut:   "2026-12-27",
		PhoneNumber: "08123456789",
	})
	var createBody model.WebResponse[*model.BookingResponse]
	data, _ := io.ReadAll(createResp.Body)
	json.Unmarshal(data, &createBody)
	bookingID := createBody.Data.ID

	db.Model(&entity.Booking{}).Where("id = ?", bookingID).Update("status", entity.StatusReview)

	adminToken := SeedAdminAndGetToken(t)

	req := httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/api/admin/bookings/%s/done", bookingID), nil)
	req.Header.Set("Cookie", "token="+adminToken)
	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var respBody model.WebResponse[*model.BookingResponse]
	body, _ := io.ReadAll(resp.Body)
	json.Unmarshal(body, &respBody)
	assert.Equal(t, "DONE", respBody.Data.Status)
}
