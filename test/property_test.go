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
	exp := SeedExperience()

	resp := doCreateProperty(t, model.PropertyCreateRequest{
		ExperienceID: exp.ID,
		Host: &model.HostCreateRequest{
			Name:        "Budi Santoso",
			Email:       "budi@example.com",
			PhoneNumber: "08123456789",
		},
		PropertyType: "homestay",
		BookingType:  "room",
	})

	body, _ := io.ReadAll(resp.Body)
	var respBody model.WebResponse[model.PropertyResponse]
	json.Unmarshal(body, &respBody)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.True(t, respBody.Success)
	assert.NotEmpty(t, respBody.Data.ID)
	assert.Equal(t, "homestay", respBody.Data.PropertyType)
	assert.Equal(t, "room", respBody.Data.BookingType)
	assert.Equal(t, "Budi Santoso", respBody.Data.Host.Name)
}

func TestCreateProperty_SuccessWithExistingHostID(t *testing.T) {
	ClearAll()
	exp := SeedExperience()
	host := SeedHostProfile()
	hostID := host.ID

	resp := doCreateProperty(t, model.PropertyCreateRequest{
		ExperienceID: exp.ID,
		HostID:       &hostID,
		PropertyType: "villa",
		BookingType:  "unit",
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

func TestCreateProperty_MissingExperienceID(t *testing.T) {
	ClearAll()
	host := SeedHostProfile()
	hostID := host.ID

	resp := doCreateProperty(t, model.PropertyCreateRequest{
		HostID:       &hostID,
		PropertyType: "homestay",
		BookingType:  "room",
	})

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestCreateProperty_MissingHostAndHostID(t *testing.T) {
	ClearAll()
	exp := SeedExperience()

	resp := doCreateProperty(t, model.PropertyCreateRequest{
		ExperienceID: exp.ID,
		PropertyType: "homestay",
		BookingType:  "room",
	})

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestCreateProperty_InlineHostMissingName(t *testing.T) {
	ClearAll()
	exp := SeedExperience()

	resp := doCreateProperty(t, model.PropertyCreateRequest{
		ExperienceID: exp.ID,
		Host: &model.HostCreateRequest{
			Email:       "budi@example.com",
			PhoneNumber: "08123456789",
		},
		PropertyType: "homestay",
	})

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestCreateProperty_HostIDNotFound(t *testing.T) {
	ClearAll()
	exp := SeedExperience()
	notExistHostID := "host-id-tidak-ada"

	resp := doCreateProperty(t, model.PropertyCreateRequest{
		ExperienceID: exp.ID,
		HostID:       &notExistHostID,
		PropertyType: "homestay",
	})

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

// ─── Get Property by Experience ID ───────────────────────────────────────────

func TestGetPropertyByExperienceID_Success(t *testing.T) {
	ClearAll()
	exp := SeedExperience()
	host := SeedHostProfile()
	SeedProperty(host.ID, exp.ID)

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/properties/%s", exp.ID), nil)
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
	assert.Equal(t, exp.ID, respBody.Data.Experience.ID)
}

func TestGetPropertyByExperienceID_NotFound(t *testing.T) {
	ClearAll()

	req := httptest.NewRequest(http.MethodGet, "/api/properties/experience-id-tidak-ada", nil)
	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}
