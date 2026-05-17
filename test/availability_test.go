package test

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"id.diengs.backend/internal/model"
)

func doCheckAvailability(t *testing.T, rentableID, checkIn, checkOut string) *http.Response {
	t.Helper()
	url := fmt.Sprintf("/api/rentables/%s/availability?check_in=%s&check_out=%s", rentableID, checkIn, checkOut)
	r := httptest.NewRequest(http.MethodGet, url, nil)
	resp, err := app.Test(r)
	assert.Nil(t, err)
	return resp
}

// ─── Check Availability ───────────────────────────────────────────────────────

func TestCheckAvailability_NoBookings(t *testing.T) {
	ClearAll()
	rentable := SeedRentableWithDeps()

	resp := doCheckAvailability(t, rentable.ID, "2026-07-01", "2026-07-04")

	body, _ := io.ReadAll(resp.Body)
	var respBody model.WebResponse[[]model.AvailabilityResponse]
	json.Unmarshal(body, &respBody)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.True(t, respBody.Success)
	assert.Len(t, respBody.Data, 3)

	for _, day := range respBody.Data {
		assert.Equal(t, rentable.Stock, day.AvailableCount)
		assert.True(t, day.IsAvailable)
	}
}

func TestCheckAvailability_AfterBooking(t *testing.T) {
	ClearAll()
	rentable := SeedRentableWithDeps()
	token := RegisterAndGetCookie(t)

	doCreateBooking(t, token, model.BookingCreateRequest{
		PropertyID: rentable.PropertyID,
		RentableID: rentable.ID,
		CheckIn:    "2026-07-01",
		CheckOut:   "2026-07-03",
		Quantity:   2,
	})

	resp := doCheckAvailability(t, rentable.ID, "2026-07-01", "2026-07-04")

	body, _ := io.ReadAll(resp.Body)
	var respBody model.WebResponse[[]model.AvailabilityResponse]
	json.Unmarshal(body, &respBody)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Len(t, respBody.Data, 3)

	for _, day := range respBody.Data {
		switch day.Date {
		case "2026-07-01", "2026-07-02":
			assert.Equal(t, rentable.Stock-2, day.AvailableCount)
			assert.True(t, day.IsAvailable)
		case "2026-07-03":
			assert.Equal(t, rentable.Stock, day.AvailableCount)
			assert.True(t, day.IsAvailable)
		}
	}
}

func TestCheckAvailability_FullyBooked(t *testing.T) {
	ClearAll()
	rentable := SeedRentableWith(5, "room")
	token := RegisterAndGetCookie(t)

	doCreateBooking(t, token, model.BookingCreateRequest{
		PropertyID: rentable.PropertyID,
		RentableID: rentable.ID,
		CheckIn:    "2026-08-01",
		CheckOut:   "2026-08-02",
		Quantity:   5,
	})

	resp := doCheckAvailability(t, rentable.ID, "2026-08-01", "2026-08-02")

	body, _ := io.ReadAll(resp.Body)
	var respBody model.WebResponse[[]model.AvailabilityResponse]
	json.Unmarshal(body, &respBody)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Len(t, respBody.Data, 1)
	assert.Equal(t, 0, respBody.Data[0].AvailableCount)
	assert.False(t, respBody.Data[0].IsAvailable)
}

func TestCheckAvailability_RentableNotFound(t *testing.T) {
	ClearAll()

	resp := doCheckAvailability(t, "rentable-id-tidak-ada", "2026-07-01", "2026-07-03")
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestCheckAvailability_InvalidDateRange(t *testing.T) {
	ClearAll()
	rentable := SeedRentableWithDeps()

	resp := doCheckAvailability(t, rentable.ID, "2026-07-03", "2026-07-01")
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestCheckAvailability_MissingParams(t *testing.T) {
	ClearAll()
	rentable := SeedRentableWithDeps()

	resp := doCheckAvailability(t, rentable.ID, "", "2026-07-03")
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}
