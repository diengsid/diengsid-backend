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

func doCreateRentable(t *testing.T, req model.RentableCreateRequest) *http.Response {
	t.Helper()
	body, _ := json.Marshal(req)
	r := httptest.NewRequest(http.MethodPost, "/api/rentables", bytes.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(r)
	assert.Nil(t, err)
	return resp
}

// ─── Create Rentable ──────────────────────────────────────────────────────────

func TestCreateRentable_Success(t *testing.T) {
	ClearAll()
	exp := SeedExperience()
	host := SeedHostProfile()
	prop := SeedProperty(host.ID, exp.ID)

	resp := doCreateRentable(t, model.RentableCreateRequest{
		PropertyID: prop.ID,
		Type:       "room",
		Name:       "Kamar Standard",
		ImageUrl:   "https://example.com/kamar.jpg",
		Capacity:   2,
		BasePrice:  200_000,
		Discount:   10,
		Stock:      5,
	})

	body, _ := io.ReadAll(resp.Body)
	var respBody model.WebResponse[model.RentableResponse]
	json.Unmarshal(body, &respBody)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.True(t, respBody.Success)
	assert.NotEmpty(t, respBody.Data.ID)
	assert.Equal(t, prop.ID, respBody.Data.PropertyID)
	assert.Equal(t, "room", respBody.Data.Type)
	assert.Equal(t, "Kamar Standard", respBody.Data.Name)
	assert.Equal(t, 200_000.0, respBody.Data.BasePrice)
	assert.Equal(t, 10.0, respBody.Data.Discount)
	assert.Equal(t, 5, respBody.Data.Stock)
}

func TestCreateRentable_SuccessUnitType(t *testing.T) {
	ClearAll()
	exp := SeedExperience()
	host := SeedHostProfile()
	prop := SeedProperty(host.ID, exp.ID)

	resp := doCreateRentable(t, model.RentableCreateRequest{
		PropertyID: prop.ID,
		Type:       "unit",
		Name:       "Vila Utama",
		BasePrice:  500_000,
		Stock:      1,
	})

	body, _ := io.ReadAll(resp.Body)
	var respBody model.WebResponse[model.RentableResponse]
	json.Unmarshal(body, &respBody)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.True(t, respBody.Success)
	assert.NotEmpty(t, respBody.Data.ID)
	assert.Equal(t, "unit", respBody.Data.Type)
	assert.Equal(t, 1, respBody.Data.Stock)
}

func TestCreateRentable_PropertyNotFound(t *testing.T) {
	ClearAll()

	resp := doCreateRentable(t, model.RentableCreateRequest{
		PropertyID: "property-id-tidak-ada",
		Type:       "room",
		Name:       "Kamar Standard",
		BasePrice:  200_000,
		Stock:      1,
	})

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestCreateRentable_UnitTypeStockGreaterThan1(t *testing.T) {
	ClearAll()
	exp := SeedExperience()
	host := SeedHostProfile()
	prop := SeedProperty(host.ID, exp.ID)

	resp := doCreateRentable(t, model.RentableCreateRequest{
		PropertyID: prop.ID,
		Type:       "unit",
		Name:       "Vila Utama",
		BasePrice:  500_000,
		Stock:      2,
	})

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}
