package test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"id.diengs.backend/internal/model"
)

func doListAmenities(t *testing.T) *http.Response {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, "/api/amenities", nil)
	resp, err := app.Test(req)
	assert.Nil(t, err)
	return resp
}

func doCreateAmenity(t *testing.T, req model.AmenityCreateRequest) *http.Response {
	t.Helper()
	body, _ := json.Marshal(req)
	r := httptest.NewRequest(http.MethodPost, "/api/amenities", bytes.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(r)
	assert.Nil(t, err)
	return resp
}

func doSetPropertyAmenities(t *testing.T, propertyID string, amenityIDs []string) *http.Response {
	t.Helper()
	body, _ := json.Marshal(model.SetAmenitiesRequest{AmenityIDs: amenityIDs})
	r := httptest.NewRequest(http.MethodPut, "/api/properties/"+propertyID+"/amenities", bytes.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(r)
	assert.Nil(t, err)
	return resp
}

func doSetRentableAmenities(t *testing.T, rentableID string, amenityIDs []string) *http.Response {
	t.Helper()
	body, _ := json.Marshal(model.SetAmenitiesRequest{AmenityIDs: amenityIDs})
	r := httptest.NewRequest(http.MethodPut, "/api/rentables/"+rentableID+"/amenities", bytes.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(r)
	assert.Nil(t, err)
	return resp
}

// ─── List Amenities ───────────────────────────────────────────────────────────

func TestListAmenities_Empty(t *testing.T) {
	ClearAll()

	resp := doListAmenities(t)
	body, _ := io.ReadAll(resp.Body)
	var respBody model.WebResponse[[]model.AmenityResponse]
	json.Unmarshal(body, &respBody)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.True(t, respBody.Success)
	assert.Empty(t, respBody.Data)
}

func TestListAmenities_ReturnsSeedData(t *testing.T) {
	ClearAll()
	SeedAmenity("WiFi", "wifi", "connectivity")
	SeedAmenity("Parkir", "parking", "facilities")

	resp := doListAmenities(t)
	body, _ := io.ReadAll(resp.Body)
	var respBody model.WebResponse[[]model.AmenityResponse]
	json.Unmarshal(body, &respBody)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.True(t, respBody.Success)
	assert.Len(t, respBody.Data, 2)
}

// ─── Create Amenity ───────────────────────────────────────────────────────────

func TestCreateAmenity_Success(t *testing.T) {
	ClearAll()

	resp := doCreateAmenity(t, model.AmenityCreateRequest{
		Name:     "WiFi",
		Icon:     "wifi",
		Category: "connectivity",
	})
	body, _ := io.ReadAll(resp.Body)
	var respBody model.WebResponse[model.AmenityResponse]
	json.Unmarshal(body, &respBody)

	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	assert.True(t, respBody.Success)
	assert.NotEmpty(t, respBody.Data.ID)
	assert.Equal(t, "WiFi", respBody.Data.Name)
	assert.Equal(t, "wifi", respBody.Data.Icon)
	assert.Equal(t, "connectivity", respBody.Data.Category)
}

func TestCreateAmenity_SuccessMinimal(t *testing.T) {
	ClearAll()

	resp := doCreateAmenity(t, model.AmenityCreateRequest{Name: "Parkir"})
	body, _ := io.ReadAll(resp.Body)
	var respBody model.WebResponse[model.AmenityResponse]
	json.Unmarshal(body, &respBody)

	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	assert.True(t, respBody.Success)
	assert.NotEmpty(t, respBody.Data.ID)
	assert.Equal(t, "Parkir", respBody.Data.Name)
}

func TestCreateAmenity_MissingName(t *testing.T) {
	ClearAll()

	resp := doCreateAmenity(t, model.AmenityCreateRequest{Icon: "wifi"})
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

// ─── Set Property Amenities ───────────────────────────────────────────────────

func TestSetPropertyAmenities_Success(t *testing.T) {
	ClearAll()
	host := SeedHostProfile()
	prop := SeedProperty(host.ID)
	a1 := SeedAmenity("WiFi", "wifi", "connectivity")
	a2 := SeedAmenity("Parkir", "parking", "facilities")

	resp := doSetPropertyAmenities(t, prop.ID, []string{a1.ID, a2.ID})
	body, _ := io.ReadAll(resp.Body)
	var respBody model.WebResponse[[]model.AmenityResponse]
	json.Unmarshal(body, &respBody)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.True(t, respBody.Success)
	assert.Len(t, respBody.Data, 2)
}

func TestSetPropertyAmenities_Replaces(t *testing.T) {
	ClearAll()
	host := SeedHostProfile()
	prop := SeedProperty(host.ID)
	a1 := SeedAmenity("WiFi", "wifi", "connectivity")
	a2 := SeedAmenity("Parkir", "parking", "facilities")

	// set awal: dua amenity
	doSetPropertyAmenities(t, prop.ID, []string{a1.ID, a2.ID})

	// ganti dengan satu amenity saja
	resp := doSetPropertyAmenities(t, prop.ID, []string{a1.ID})
	body, _ := io.ReadAll(resp.Body)
	var respBody model.WebResponse[[]model.AmenityResponse]
	json.Unmarshal(body, &respBody)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Len(t, respBody.Data, 1)
	assert.Equal(t, a1.ID, respBody.Data[0].ID)
}

func TestSetPropertyAmenities_ClearAll(t *testing.T) {
	ClearAll()
	host := SeedHostProfile()
	prop := SeedProperty(host.ID)
	a1 := SeedAmenity("WiFi", "wifi", "connectivity")
	doSetPropertyAmenities(t, prop.ID, []string{a1.ID})

	// kosongkan
	resp := doSetPropertyAmenities(t, prop.ID, []string{})
	body, _ := io.ReadAll(resp.Body)
	var respBody model.WebResponse[[]model.AmenityResponse]
	json.Unmarshal(body, &respBody)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Empty(t, respBody.Data)
}

func TestSetPropertyAmenities_PropertyNotFound(t *testing.T) {
	ClearAll()

	resp := doSetPropertyAmenities(t, "property-id-tidak-ada", []string{})
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

// ─── Set Rentable Amenities ───────────────────────────────────────────────────

func TestSetRentableAmenities_Success(t *testing.T) {
	ClearAll()
	rentable := SeedRentableWithDeps()
	a1 := SeedAmenity("AC", "air-conditioning", "comfort")
	a2 := SeedAmenity("Air Panas", "shower", "bathroom")

	resp := doSetRentableAmenities(t, rentable.ID, []string{a1.ID, a2.ID})
	body, _ := io.ReadAll(resp.Body)
	var respBody model.WebResponse[[]model.AmenityResponse]
	json.Unmarshal(body, &respBody)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.True(t, respBody.Success)
	assert.Len(t, respBody.Data, 2)
}

func TestSetRentableAmenities_Replaces(t *testing.T) {
	ClearAll()
	rentable := SeedRentableWithDeps()
	a1 := SeedAmenity("AC", "air-conditioning", "comfort")
	a2 := SeedAmenity("Air Panas", "shower", "bathroom")

	doSetRentableAmenities(t, rentable.ID, []string{a1.ID, a2.ID})

	resp := doSetRentableAmenities(t, rentable.ID, []string{a2.ID})
	body, _ := io.ReadAll(resp.Body)
	var respBody model.WebResponse[[]model.AmenityResponse]
	json.Unmarshal(body, &respBody)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Len(t, respBody.Data, 1)
	assert.Equal(t, a2.ID, respBody.Data[0].ID)
}

func TestSetRentableAmenities_ClearAll(t *testing.T) {
	ClearAll()
	rentable := SeedRentableWithDeps()
	a1 := SeedAmenity("AC", "air-conditioning", "comfort")
	doSetRentableAmenities(t, rentable.ID, []string{a1.ID})

	resp := doSetRentableAmenities(t, rentable.ID, []string{})
	body, _ := io.ReadAll(resp.Body)
	var respBody model.WebResponse[[]model.AmenityResponse]
	json.Unmarshal(body, &respBody)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Empty(t, respBody.Data)
}

func TestSetRentableAmenities_RentableNotFound(t *testing.T) {
	ClearAll()

	resp := doSetRentableAmenities(t, "rentable-id-tidak-ada", []string{})
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}
