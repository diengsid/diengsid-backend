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
	"id.diengs.backend/internal/model"
)

func doCreateProperty(t *testing.T, req model.PropertyCreateRequest) *http.Response {
	t.Helper()
	body, _ := json.Marshal(req)
	r := httptest.NewRequest(http.MethodPost, "/api/properties", bytes.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(r)
	assert.Nil(t, err)
	return resp
}

// ─── Create Property ──────────────────────────────────────────────────────────

func TestCreateProperty_SuccessWithNewHost(t *testing.T) {
	ClearAll()

	resp := doCreateProperty(t, model.PropertyCreateRequest{
		Host: &model.HostCreateRequest{
			Name:        "Budi Santoso",
			Email:       "budi@example.com",
			PhoneNumber: "08123456789",
		},
		PropertyType: "homestay",
		BookingType:  "room",
		Title:        "Homestay Dieng Sejuk",
		Address:      "Dieng, Wonosobo",
		Description:  "Homestay nyaman dengan pemandangan alam Dieng.",
	})

	body, _ := io.ReadAll(resp.Body)
	var respBody model.WebResponse[model.PropertyResponse]
	json.Unmarshal(body, &respBody)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.True(t, respBody.Success)
	assert.NotEmpty(t, respBody.Data.ID)
	assert.Equal(t, "homestay", respBody.Data.PropertyType)
	assert.Equal(t, "room", respBody.Data.BookingType)
	assert.Equal(t, "Homestay Dieng Sejuk", respBody.Data.Title)
	assert.Equal(t, "Budi Santoso", respBody.Data.Host.Name)
}

func TestCreateProperty_SuccessWithExistingHostID(t *testing.T) {
	ClearAll()
	host := SeedHostProfile()
	hostID := host.ID

	resp := doCreateProperty(t, model.PropertyCreateRequest{
		HostID:       &hostID,
		PropertyType: "villa",
		BookingType:  "unit",
		Title:        "Villa Dieng Indah",
		Address:      "Dieng, Wonosobo",
		Description:  "Villa dengan fasilitas lengkap di kawasan Dieng.",
	})

	body, _ := io.ReadAll(resp.Body)
	var respBody model.WebResponse[model.PropertyResponse]
	json.Unmarshal(body, &respBody)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.True(t, respBody.Success)
	assert.NotEmpty(t, respBody.Data.ID)
	assert.Equal(t, "villa", respBody.Data.PropertyType)
	assert.Equal(t, host.Name, respBody.Data.Host.Name)
}

func TestCreateProperty_SuccessWithImages(t *testing.T) {
	ClearAll()

	thumbnailURL := "https://example.com/thumb.jpg"
	lat, lng := -7.2132, 109.9199

	resp := doCreateProperty(t, model.PropertyCreateRequest{
		Host: &model.HostCreateRequest{
			Name:        "Sari",
			Email:       "sari@example.com",
			PhoneNumber: "08111111111",
		},
		PropertyType: "guesthouse",
		BookingType:  "room",
		Title:        "Guest House Sikunir",
		Address:      "Sikunir, Kejajar, Wonosobo",
		Description:  "Penginapan dekat spot sunrise Sikunir.",
		ThumbnailURL: &thumbnailURL,
		Lat:          &lat,
		Lng:          &lng,
		Images: []model.PropertyImageCreateRequest{
			{ImageURL: "https://example.com/img1.jpg", IsPrimary: true},
			{ImageURL: "https://example.com/img2.jpg", IsPrimary: false},
		},
	})

	body, _ := io.ReadAll(resp.Body)
	var respBody model.WebResponse[model.PropertyResponse]
	json.Unmarshal(body, &respBody)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.True(t, respBody.Success)
	assert.Len(t, respBody.Data.Images, 2)
	assert.Equal(t, thumbnailURL, *respBody.Data.ThumbnailURL)
}

func TestCreateProperty_MissingTitle(t *testing.T) {
	ClearAll()
	host := SeedHostProfile()
	hostID := host.ID

	resp := doCreateProperty(t, model.PropertyCreateRequest{
		HostID:       &hostID,
		PropertyType: "homestay",
		BookingType:  "room",
		Address:      "Dieng, Wonosobo",
		Description:  "Deskripsi properti.",
	})

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestCreateProperty_MissingAddress(t *testing.T) {
	ClearAll()
	host := SeedHostProfile()
	hostID := host.ID

	resp := doCreateProperty(t, model.PropertyCreateRequest{
		HostID:       &hostID,
		PropertyType: "homestay",
		BookingType:  "room",
		Title:        "Homestay Dieng",
		Description:  "Deskripsi properti.",
	})

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestCreateProperty_MissingDescription(t *testing.T) {
	ClearAll()
	host := SeedHostProfile()
	hostID := host.ID

	resp := doCreateProperty(t, model.PropertyCreateRequest{
		HostID:       &hostID,
		PropertyType: "homestay",
		BookingType:  "room",
		Title:        "Homestay Dieng",
		Address:      "Dieng, Wonosobo",
	})

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestCreateProperty_MissingHostAndHostID(t *testing.T) {
	ClearAll()

	resp := doCreateProperty(t, model.PropertyCreateRequest{
		PropertyType: "homestay",
		BookingType:  "room",
		Title:        "Homestay Dieng",
		Address:      "Dieng, Wonosobo",
		Description:  "Deskripsi properti.",
	})

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestCreateProperty_InlineHostMissingName(t *testing.T) {
	ClearAll()

	resp := doCreateProperty(t, model.PropertyCreateRequest{
		Host: &model.HostCreateRequest{
			Email:       "budi@example.com",
			PhoneNumber: "08123456789",
		},
		PropertyType: "homestay",
		Title:        "Homestay Dieng",
		Address:      "Dieng, Wonosobo",
		Description:  "Deskripsi properti.",
	})

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestCreateProperty_HostIDNotFound(t *testing.T) {
	ClearAll()
	notExistHostID := "host-id-tidak-ada"

	resp := doCreateProperty(t, model.PropertyCreateRequest{
		HostID:       &notExistHostID,
		PropertyType: "homestay",
		Title:        "Homestay Dieng",
		Address:      "Dieng, Wonosobo",
		Description:  "Deskripsi properti.",
	})

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

// ─── Get Property by ID ───────────────────────────────────────────────────────

func TestGetPropertyByID_Success(t *testing.T) {
	ClearAll()
	host := SeedHostProfile()
	prop := SeedProperty(host.ID)

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/properties/%s", prop.ID), nil)
	resp, err := app.Test(req)

	assert.Nil(t, err)
	body, _ := io.ReadAll(resp.Body)
	var respBody model.WebResponse[model.PropertyResponse]
	json.Unmarshal(body, &respBody)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.True(t, respBody.Success)
	assert.NotEmpty(t, respBody.Data.ID)
	assert.Equal(t, "homestay", respBody.Data.PropertyType)
	assert.Equal(t, "room", respBody.Data.BookingType)
	assert.Equal(t, "Homestay Dieng", respBody.Data.Title)
}

func TestGetPropertyByID_NotFound(t *testing.T) {
	ClearAll()

	req := httptest.NewRequest(http.MethodGet, "/api/properties/property-id-tidak-ada", nil)
	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}
