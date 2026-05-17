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

func doCreateExperience(t *testing.T, req model.ExperienceCreateRequest) *http.Response {
	t.Helper()
	body, _ := json.Marshal(req)
	r := httptest.NewRequest(http.MethodPost, "/api/experiences", bytes.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(r)
	assert.Nil(t, err)
	return resp
}

// ─── Create Experience ────────────────────────────────────────────────────────

func TestCreateExperience_Success(t *testing.T) {
	ClearAll()

	thumbnailURL := "https://example.com/thumb.jpg"
	lat, lng := -7.2132, 109.9199

	resp := doCreateExperience(t, model.ExperienceCreateRequest{
		ExperienceType: "nature",
		Title:          "Sunrise di Sikunir",
		Address:        "Sikunir, Kejajar, Wonosobo",
		Description:    "Nikmati matahari terbit dari puncak Sikunir.",
		ThumbnailURL:   &thumbnailURL,
		Lat:            &lat,
		Lng:            &lng,
		BasePrice:      150_000,
		Images: []model.ExperienceCreateImageRequest{
			{ImageURL: "https://example.com/img1.jpg", IsPrimary: true},
			{ImageURL: "https://example.com/img2.jpg", IsPrimary: false},
		},
	})

	body, _ := io.ReadAll(resp.Body)
	var respBody model.WebResponse[model.ExperienceResponse]
	json.Unmarshal(body, &respBody)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.True(t, respBody.Success)
	assert.NotEmpty(t, respBody.Data.ID)
	assert.Equal(t, "nature", respBody.Data.ExperienceType)
	assert.Equal(t, "Sunrise di Sikunir", respBody.Data.Title)
	assert.Equal(t, 150_000.0, respBody.Data.BasePrice)
	assert.Len(t, respBody.Data.Images, 2)
}

func TestCreateExperience_SuccessMinimal(t *testing.T) {
	ClearAll()

	resp := doCreateExperience(t, model.ExperienceCreateRequest{
		ExperienceType: "culinary",
		Title:          "Tour Kuliner Dieng",
		Address:        "Pasar Dieng, Banjarnegara",
		Description:    "Jelajahi kuliner khas Dieng.",
		BasePrice:      75_000,
	})

	body, _ := io.ReadAll(resp.Body)
	var respBody model.WebResponse[model.ExperienceResponse]
	json.Unmarshal(body, &respBody)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.True(t, respBody.Success)
	assert.NotEmpty(t, respBody.Data.ID)
	assert.Empty(t, respBody.Data.Images)
}

func TestCreateExperience_MissingExperienceType(t *testing.T) {
	ClearAll()

	resp := doCreateExperience(t, model.ExperienceCreateRequest{
		Title:       "Telaga Warna",
		Address:     "Dieng, Banjarnegara",
		Description: "Danau warna-warni.",
		BasePrice:   50_000,
	})

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestCreateExperience_MissingTitle(t *testing.T) {
	ClearAll()

	resp := doCreateExperience(t, model.ExperienceCreateRequest{
		ExperienceType: "nature",
		Address:        "Dieng, Banjarnegara",
		Description:    "Danau warna-warni.",
		BasePrice:      50_000,
	})

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestCreateExperience_MissingAddress(t *testing.T) {
	ClearAll()

	resp := doCreateExperience(t, model.ExperienceCreateRequest{
		ExperienceType: "nature",
		Title:          "Telaga Warna",
		Description:    "Danau warna-warni.",
		BasePrice:      50_000,
	})

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestCreateExperience_MissingDescription(t *testing.T) {
	ClearAll()

	resp := doCreateExperience(t, model.ExperienceCreateRequest{
		ExperienceType: "nature",
		Title:          "Telaga Warna",
		Address:        "Dieng, Banjarnegara",
		BasePrice:      50_000,
	})

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestCreateExperience_MissingBasePrice(t *testing.T) {
	ClearAll()

	resp := doCreateExperience(t, model.ExperienceCreateRequest{
		ExperienceType: "nature",
		Title:          "Telaga Warna",
		Address:        "Dieng, Banjarnegara",
		Description:    "Danau warna-warni.",
	})

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

// ─── Search Experience ────────────────────────────────────────────────────────

func TestSearchExperience_Success(t *testing.T) {
	ClearAll()
	SeedExperienceWith("nature", "Sikunir Sunrise Tour", 150_000)
	SeedExperienceWith("culinary", "Tour Kuliner Dieng", 75_000)

	req := httptest.NewRequest(http.MethodGet, "/api/experiences", nil)
	resp, err := app.Test(req)

	assert.Nil(t, err)
	body, _ := io.ReadAll(resp.Body)
	var respBody model.WebResponse[[]model.ExperienceResponse]
	json.Unmarshal(body, &respBody)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Len(t, respBody.Data, 2)
	assert.NotNil(t, respBody.Paging)
	assert.Equal(t, int64(2), respBody.Paging.TotalItem)
}

func TestSearchExperience_FilterByKeyword(t *testing.T) {
	ClearAll()
	SeedExperienceWith("nature", "Sikunir Sunrise Tour", 150_000)
	SeedExperienceWith("culinary", "Tour Kuliner Dieng", 75_000)

	req := httptest.NewRequest(http.MethodGet, "/api/experiences?key=sikunir", nil)
	resp, err := app.Test(req)

	assert.Nil(t, err)
	body, _ := io.ReadAll(resp.Body)
	var respBody model.WebResponse[[]model.ExperienceResponse]
	json.Unmarshal(body, &respBody)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Len(t, respBody.Data, 1)
	assert.Contains(t, respBody.Data[0].Title, "Sikunir")
}

func TestSearchExperience_FilterByType(t *testing.T) {
	ClearAll()
	SeedExperienceWith("nature", "Sikunir Sunrise Tour", 150_000)
	SeedExperienceWith("nature", "Telaga Warna", 50_000)
	SeedExperienceWith("culinary", "Tour Kuliner Dieng", 75_000)

	req := httptest.NewRequest(http.MethodGet, "/api/experiences?type=nature", nil)
	resp, err := app.Test(req)

	assert.Nil(t, err)
	body, _ := io.ReadAll(resp.Body)
	var respBody model.WebResponse[[]model.ExperienceResponse]
	json.Unmarshal(body, &respBody)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Len(t, respBody.Data, 2)
	assert.Equal(t, int64(2), respBody.Paging.TotalItem)
}

func TestSearchExperience_EmptyResult(t *testing.T) {
	ClearAll()
	SeedExperienceWith("nature", "Sikunir Sunrise Tour", 150_000)

	req := httptest.NewRequest(http.MethodGet, "/api/experiences?key=tidakadahasilnya", nil)
	resp, err := app.Test(req)

	assert.Nil(t, err)
	body, _ := io.ReadAll(resp.Body)
	var respBody model.WebResponse[[]model.ExperienceResponse]
	json.Unmarshal(body, &respBody)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Empty(t, respBody.Data)
	assert.Equal(t, int64(0), respBody.Paging.TotalItem)
}
