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

func cookieFromResp(resp *http.Response, name string) string {
	for _, c := range resp.Cookies() {
		if c.Name == name {
			return c.Value
		}
	}
	return ""
}

// ─── Register ─────────────────────────────────────────────────────────────────

func TestRegister_Success(t *testing.T) {
	ClearAll()

	resp := DoRegister(t, "Ahmad Rifai", "ahmad@example.com", "password123")

	body, _ := io.ReadAll(resp.Body)
	var respBody model.WebResponse[*model.AuthResponse]
	json.Unmarshal(body, &respBody)

	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	assert.True(t, respBody.Success)
	assert.Equal(t, "ahmad@example.com", respBody.Data.User.Email)
	assert.Equal(t, "Ahmad Rifai", respBody.Data.User.Name)
	assert.False(t, respBody.Data.User.EmailVerified)
	assert.NotEmpty(t, cookieFromResp(resp, "token"))
}

func TestRegister_DuplicateEmail(t *testing.T) {
	ClearAll()
	DoRegister(t, "Ahmad Rifai", "ahmad@example.com", "password123")

	resp := DoRegister(t, "User Lain", "ahmad@example.com", "password456")

	assert.Equal(t, http.StatusConflict, resp.StatusCode)
}

func TestRegister_PasswordTooShort(t *testing.T) {
	ClearAll()

	resp := DoRegister(t, "Ahmad", "ahmad@example.com", "short")
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestRegister_MissingName(t *testing.T) {
	ClearAll()

	body, _ := json.Marshal(map[string]string{
		"email":    "ahmad@example.com",
		"password": "password123",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)

	assert.Nil(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestRegister_InvalidEmail(t *testing.T) {
	ClearAll()

	resp := DoRegister(t, "Ahmad", "bukan-email", "password123")
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

// ─── Login ────────────────────────────────────────────────────────────────────

func TestLogin_Success(t *testing.T) {
	ClearAll()
	DoRegister(t, "Ahmad Rifai", "ahmad@example.com", "password123")

	body, _ := json.Marshal(model.LoginRequest{
		Email:    "ahmad@example.com",
		Password: "password123",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)

	assert.Nil(t, err)

	body2, _ := io.ReadAll(resp.Body)
	var respBody model.WebResponse[*model.AuthResponse]
	json.Unmarshal(body2, &respBody)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.True(t, respBody.Success)
	assert.Equal(t, "ahmad@example.com", respBody.Data.User.Email)
	assert.NotEmpty(t, cookieFromResp(resp, "token"))
}

func TestLogin_WrongPassword(t *testing.T) {
	ClearAll()
	DoRegister(t, "Ahmad Rifai", "ahmad@example.com", "password123")

	body, _ := json.Marshal(model.LoginRequest{
		Email:    "ahmad@example.com",
		Password: "wrongpassword",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestLogin_EmailNotFound(t *testing.T) {
	ClearAll()

	body, _ := json.Marshal(model.LoginRequest{
		Email:    "tidakada@example.com",
		Password: "password123",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestLogin_ValidationError(t *testing.T) {
	ClearAll()

	body, _ := json.Marshal(map[string]string{
		"email": "bukan-email",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}
