package test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"id.diengs.backend/internal/entity"
	"id.diengs.backend/internal/model"
)

// ─── Clear helpers ────────────────────────────────────────────────────────────

func ClearAll() {
	ClearPayments()
	ClearBookings()
	ClearAvailabilities()
	ClearSessions()
	ClearRentableAmenities()
	ClearPropertyAmenities()
	ClearRentables()
	ClearProperties()
	ClearHostProfiles()
	ClearExperienceImages()
	ClearExperiences()
	ClearUsers()
	ClearAmenities()
}

func ClearBookings() {
	if err := db.Where("id IS NOT NULL").Delete(&entity.Booking{}).Error; err != nil {
		log.Fatalf("failed to clear bookings: %v", err)
	}
}

func ClearAvailabilities() {
	if err := db.Where("id IS NOT NULL").Delete(&entity.Availability{}).Error; err != nil {
		log.Fatalf("failed to clear availabilities: %v", err)
	}
}

func ClearSessions() {
	if err := db.Where("id IS NOT NULL").Delete(&entity.Session{}).Error; err != nil {
		log.Fatalf("failed to clear sessions: %v", err)
	}
}

func ClearRentables() {
	if err := db.Where("id IS NOT NULL").Delete(&entity.Rentable{}).Error; err != nil {
		log.Fatalf("failed to clear rentables: %v", err)
	}
}

func ClearProperties() {
	if err := db.Where("id IS NOT NULL").Delete(&entity.Property{}).Error; err != nil {
		log.Fatalf("failed to clear properties: %v", err)
	}
}

func ClearHostProfiles() {
	if err := db.Where("id IS NOT NULL").Delete(&entity.HostProfile{}).Error; err != nil {
		log.Fatalf("failed to clear host_profiles: %v", err)
	}
}

func ClearExperienceImages() {
	if err := db.Where("id IS NOT NULL").Delete(&entity.ExperienceImage{}).Error; err != nil {
		log.Fatalf("failed to clear experience_images: %v", err)
	}
}

func ClearExperiences() {
	if err := db.Where("id IS NOT NULL").Delete(&entity.Experience{}).Error; err != nil {
		log.Fatalf("failed to clear experiences: %v", err)
	}
}

func ClearUsers() {
	if err := db.Where("id IS NOT NULL").Delete(&entity.User{}).Error; err != nil {
		log.Fatalf("failed to clear users: %v", err)
	}
}

// ─── Seed helpers ─────────────────────────────────────────────────────────────

func SeedExperience() *entity.Experience {
	return SeedExperienceWith("nature", "Dieng Plateau Tour", 150_000)
}

func SeedExperienceWith(experienceType, title string, basePrice float64) *entity.Experience {
	exp := &entity.Experience{
		ExperienceType: experienceType,
		Title:          title,
		Address:        "Dieng, Wonosobo",
		Description:    "Deskripsi singkat untuk " + title,
		BasePrice:      basePrice,
	}
	if err := db.Create(exp).Error; err != nil {
		log.Fatalf("failed to seed experience: %v", err)
	}
	return exp
}

func SeedHostProfile() *entity.HostProfile {
	host := &entity.HostProfile{
		Name:        "Host Test",
		Email:       "host@test.com",
		PhoneNumber: "08123456789",
	}
	if err := db.Create(host).Error; err != nil {
		log.Fatalf("failed to seed host_profile: %v", err)
	}
	return host
}

func SeedProperty(hostID, experienceID string) *entity.Property {
	prop := &entity.Property{
		HostID:       hostID,
		ExperienceID: experienceID,
		PropertyType: "homestay",
		BookingType:  "room",
	}
	if err := db.Create(prop).Error; err != nil {
		log.Fatalf("failed to seed property: %v", err)
	}
	return prop
}

func SeedRentable(propertyID string) *entity.Rentable {
	rentable := &entity.Rentable{
		PropertyID: propertyID,
		Type:       "room",
		Name:       "Kamar Standard",
		BasePrice:  200_000,
		Discount:   10,
		Stock:      5,
	}
	if err := db.Create(rentable).Error; err != nil {
		log.Fatalf("failed to seed rentable: %v", err)
	}
	return rentable
}

func SeedRentableWithDeps() *entity.Rentable {
	exp := SeedExperience()
	host := SeedHostProfile()
	prop := SeedProperty(host.ID, exp.ID)
	return SeedRentable(prop.ID)
}

func SeedRentableWith(stock int, rentableType string) *entity.Rentable {
	exp := SeedExperience()
	host := SeedHostProfile()
	prop := SeedProperty(host.ID, exp.ID)
	rentable := &entity.Rentable{
		PropertyID: prop.ID,
		Type:       rentableType,
		Name:       "Test Rentable",
		BasePrice:  200_000,
		Discount:   0,
		Stock:      stock,
	}
	if err := db.Create(rentable).Error; err != nil {
		log.Fatalf("failed to seed rentable: %v", err)
	}
	return rentable
}

// ─── Auth helpers ─────────────────────────────────────────────────────────────

func DoRegister(t *testing.T, name, email, password string) *http.Response {
	t.Helper()
	body, _ := json.Marshal(model.RegisterRequest{
		Name:     name,
		Email:    email,
		Password: password,
	})
	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.Nil(t, err)
	return resp
}

func RegisterAndGetCookie(t *testing.T) string {
	t.Helper()
	resp := DoRegister(t, "Test User", "test@example.com", "password123")
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	for _, c := range resp.Cookies() {
		if c.Name == "token" {
			assert.NotEmpty(t, c.Value)
			return c.Value
		}
	}
	t.Fatal("token cookie not found in register response")
	return ""
}

// RegisterWithEmail registers a unique user with the given email and returns the session token.
func RegisterWithEmail(t *testing.T, name, email string) string {
	t.Helper()
	resp := DoRegister(t, name, email, "password123")
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	for _, c := range resp.Cookies() {
		if c.Name == "token" {
			return c.Value
		}
	}
	t.Fatal("token cookie not found")
	return ""
}

// SeedAdminAndGetToken inserts an admin user + session directly into the DB and returns the raw session token.
func SeedAdminAndGetToken(t *testing.T) string {
	t.Helper()

	adminUser := &entity.User{
		Name:          fmt.Sprintf("Admin-%d", time.Now().UnixNano()),
		Email:         fmt.Sprintf("admin-%d@test.com", time.Now().UnixNano()),
		Password:      "unused",
		Role:          "ADMIN",
		EmailVerified: true,
	}
	if err := db.Create(adminUser).Error; err != nil {
		t.Fatalf("failed to create admin user: %v", err)
	}

	tokenVal := uuid.NewString()
	session := &entity.Session{
		UserID:    adminUser.ID,
		Token:     tokenVal,
		ExpiredAt: time.Now().Add(24 * time.Hour).UnixMilli(),
	}
	if err := db.Create(session).Error; err != nil {
		t.Fatalf("failed to create admin session: %v", err)
	}

	return tokenVal
}

// SeedBookingInStatus creates a booking row directly with the given status (bypasses business logic).
func SeedBookingInStatus(userID string, rentable *entity.Rentable, status entity.BookingStatus) *entity.Booking {
	now := time.Now()
	booking := &entity.Booking{
		UserID:        userID,
		PropertyID:    rentable.PropertyID,
		RentableID:    rentable.ID,
		Quantity:      1,
		GuestCount:    1,
		CheckIn:       now.Add(24 * time.Hour),
		CheckOut:      now.Add(72 * time.Hour),
		TotalNight:    2,
		TotalPrice:    400_000,
		Status:        status,
		PaymentStatus: entity.PaymentUnpaid,
	}
	if err := db.Create(booking).Error; err != nil {
		log.Fatalf("failed to seed booking: %v", err)
	}
	return booking
}

// ClearPayments removes all rows from the payments table.
func ClearPayments() {
	if err := db.Where("id IS NOT NULL").Delete(&entity.Payment{}).Error; err != nil {
		log.Fatalf("failed to clear payments: %v", err)
	}
}

func ClearRentableAmenities() {
	if err := db.Exec("DELETE FROM rentable_amenities").Error; err != nil {
		log.Fatalf("failed to clear rentable_amenities: %v", err)
	}
}

func ClearPropertyAmenities() {
	if err := db.Exec("DELETE FROM property_amenities").Error; err != nil {
		log.Fatalf("failed to clear property_amenities: %v", err)
	}
}

func ClearAmenities() {
	if err := db.Where("id IS NOT NULL").Delete(&entity.Amenity{}).Error; err != nil {
		log.Fatalf("failed to clear amenities: %v", err)
	}
}

func SeedAmenity(name, icon, category string) *entity.Amenity {
	a := &entity.Amenity{Name: name, Icon: icon, Category: category}
	if err := db.Create(a).Error; err != nil {
		log.Fatalf("failed to seed amenity: %v", err)
	}
	return a
}
