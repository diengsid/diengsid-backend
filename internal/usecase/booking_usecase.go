package usecase

import (
	"context"
	"math"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"id.diengs.backend/internal/entity"
	"id.diengs.backend/internal/model"
	"id.diengs.backend/internal/pkg"
	"id.diengs.backend/internal/pkg/mailview"
	"id.diengs.backend/internal/pkg/message"
	"id.diengs.backend/internal/repository"
)

type BookingUseCase struct {
	DB               *gorm.DB
	Log              *logrus.Logger
	Validate         *validator.Validate
	BookingRepo      *repository.BookingRepo
	RentableRepo     *repository.RentableRepo
	AvailabilityRepo *repository.AvailabilityRepo
	PropertyRepo     *repository.PropertyRepo
	HostProfileRepo  *repository.HostProfileRepo
	UserRepo         *repository.UserRepo
	WA               pkg.WhatsAppSender
	Mail             *pkg.Mail
}

func NewBookingUseCase(
	db *gorm.DB,
	log *logrus.Logger,
	validate *validator.Validate,
	bookingRepo *repository.BookingRepo,
	rentableRepo *repository.RentableRepo,
	availabilityRepo *repository.AvailabilityRepo,
	propertyRepo *repository.PropertyRepo,
	hostProfileRepo *repository.HostProfileRepo,
	userRepo *repository.UserRepo,
	wa pkg.WhatsAppSender,
	mail *pkg.Mail,
) *BookingUseCase {
	return &BookingUseCase{
		DB:               db,
		Log:              log,
		Validate:         validate,
		BookingRepo:      bookingRepo,
		RentableRepo:     rentableRepo,
		AvailabilityRepo: availabilityRepo,
		PropertyRepo:     propertyRepo,
		HostProfileRepo:  hostProfileRepo,
		UserRepo:         userRepo,
		WA:               wa,
		Mail:             mail,
	}
}

// sendWA sends a WhatsApp message in a goroutine, logging any failure as a warning.
func (u *BookingUseCase) sendWA(phone, message string) {
	if phone == "" || u.WA == nil {
		return
	}
	go func() {
		if err := u.WA.SendOne(phone, message); err != nil {
			u.Log.WithError(err).Warn("failed to send whatsapp notification")
		}
	}()
}

// sendMail sends an email in a goroutine, logging any failure as a warning.
func (u *BookingUseCase) sendMail(to, subject, body string) {
	if to == "" || u.Mail == nil {
		return
	}
	go func() {
		if err := u.Mail.SendMail([]string{to}, subject, body); err != nil {
			u.Log.WithError(err).Warn("failed to send email notification")
		}
	}()
}

// notifyBookingCreated sends WA and email to both customer and host after booking is created.
func (u *BookingUseCase) notifyBookingCreated(booking *entity.Booking) {
	db := u.DB
	checkIn := booking.CheckIn.Format("02 Jan 2006")
	checkOut := booking.CheckOut.Format("02 Jan 2006")

	// customer
	user := new(entity.User)
	if err := u.UserRepo.FindById(db, user, booking.UserID); err == nil {
		u.sendWA(user.PhoneNumber, message.BookingCreatedCustomer(user.Name, booking.ID, checkIn, checkOut, booking.TotalNight, booking.GuestCount, booking.TotalPrice))
		u.sendMail(user.Email, "Booking Berhasil Dibuat - Diengs.id",
			mailview.BookingCreatedCustomerMailView(user.Name, booking.ID, "", checkIn, checkOut, booking.TotalNight, booking.GuestCount, booking.TotalPrice),
		)
	}

	// host
	property := new(entity.Property)
	if err := u.PropertyRepo.FindById(db, property, booking.PropertyID, "Host"); err == nil {
		u.sendWA(property.Host.PhoneNumber, message.BookingCreatedHost(property.Host.Name, booking.ID, property.Title, checkIn, checkOut, booking.TotalNight, booking.GuestCount, booking.TotalPrice))
		u.sendMail(property.Host.Email, "Booking Baru Masuk - Diengs.id",
			mailview.BookingCreatedHostMailView(property.Host.Name, booking.ID, property.Title, checkIn, checkOut, booking.TotalNight, booking.GuestCount, booking.TotalPrice),
		)
	}
}

// notifyBookingConfirmed sends WA and email to customer after host confirms (WAITING_PAYMENT or UNAVAILABLE).
func (u *BookingUseCase) notifyBookingConfirmed(booking *entity.Booking) {
	db := u.DB
	checkIn := booking.CheckIn.Format("02 Jan 2006")
	checkOut := booking.CheckOut.Format("02 Jan 2006")

	user := new(entity.User)
	if err := u.UserRepo.FindById(db, user, booking.UserID); err != nil {
		return
	}

	var msg, subject string
	var approved bool
	switch booking.Status {
	case entity.StatusWaiting:
		approved = true
		subject = "Booking Anda Dikonfirmasi - Diengs.id"
		msg = message.BookingConfirmedCustomer(user.Name, booking.ID, checkIn, checkOut, booking.TotalPrice)
	case entity.StatusUnavailable:
		approved = false
		subject = "Kamar Tidak Tersedia - Diengs.id"
		msg = message.BookingUnavailableCustomer(user.Name, booking.ID, checkIn, checkOut)
	default:
		return
	}

	u.sendWA(user.PhoneNumber, msg)
	u.sendMail(user.Email, subject,
		mailview.BookingConfirmedMailView(user.Name, booking.ID, checkIn, checkOut, booking.TotalPrice, approved),
	)
}

func (u *BookingUseCase) Create(ctx context.Context, userID string, req *model.BookingCreateRequest) (*model.BookingResponse, error) {
	tx := u.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	if err := u.Validate.Struct(req); err != nil {
		u.Log.WithError(err).Error("FAILED TO VALIDATE REQUEST.")
		return nil, fiber.ErrBadRequest
	}

	checkIn, err := time.Parse("2006-01-02", req.CheckIn)
	if err != nil {
		u.Log.WithError(err).Error("INVALID CHECK-IN DATE FORMAT.")
		return nil, fiber.ErrBadRequest
	}

	checkOut, err := time.Parse("2006-01-02", req.CheckOut)
	if err != nil {
		u.Log.WithError(err).Error("INVALID CHECK-OUT DATE FORMAT.")
		return nil, fiber.ErrBadRequest
	}

	if !checkOut.After(checkIn) {
		u.Log.Error("CHECK-OUT MUST BE AFTER CHECK-IN.")
		return nil, fiber.ErrBadRequest
	}

	rentable := new(entity.Rentable)
	if err := u.RentableRepo.FindById(tx, rentable, req.RentableID); err != nil {
		if err == gorm.ErrRecordNotFound {
			u.Log.WithError(err).Error("RENTABLE NOT FOUND.")
			return nil, fiber.ErrNotFound
		}
		u.Log.WithError(err).Error("FAILED TO GET RENTABLE.")
		return nil, fiber.ErrInternalServerError
	}

	quantity := req.Quantity
	if quantity <= 0 {
		quantity = 1
	}

	// ── Cek availability untuk setiap malam ──────────────────────────────────
	dates := datesInRange(checkIn, checkOut)

	var avails []entity.Availability
	if err := u.AvailabilityRepo.FindByRentableAndDateRange(tx, &avails, req.RentableID, checkIn, checkOut); err != nil {
		u.Log.WithError(err).Error("FAILED TO GET AVAILABILITY.")
		return nil, fiber.ErrInternalServerError
	}

	availMap := make(map[string]int)
	for _, a := range avails {
		availMap[time.Unix(a.Date, 0).UTC().Format("2006-01-02")] = a.AvailableCount
	}

	for _, date := range dates {
		dateStr := date.Format("2006-01-02")
		count, exists := availMap[dateStr]
		if !exists {
			count = rentable.Stock
		}
		if count < quantity {
			u.Log.Errorf("NOT ENOUGH AVAILABILITY FOR %s (have %d, need %d).", dateStr, count, quantity)
			return nil, fiber.NewError(fiber.StatusConflict, "not enough availability for "+dateStr)
		}
	}

	// ── Hitung harga ─────────────────────────────────────────────────────────
	totalNight := int(math.Round(checkOut.Sub(checkIn).Hours() / 24))
	discountAmount := rentable.BasePrice * (rentable.Discount / 100)
	pricePerNight := rentable.BasePrice - discountAmount
	totalPrice := pricePerNight * float64(totalNight) * float64(quantity)

	var firstPayment *entity.FirstPayment
	if req.FirstPayment != "" {
		fp := entity.FirstPayment(req.FirstPayment)
		firstPayment = &fp
	}

	guestCount := req.GuestCount
	if guestCount <= 0 {
		guestCount = 1
	}

	booking := &entity.Booking{
		UserID:        userID,
		PropertyID:    req.PropertyID,
		RentableID:    req.RentableID,
		Quantity:      quantity,
		GuestCount:    guestCount,
		CheckIn:       checkIn,
		CheckOut:      checkOut,
		TotalNight:    totalNight,
		TotalPrice:    totalPrice,
		Discount:      rentable.Discount,
		Status:        entity.StatusPending,
		PaymentStatus: entity.PaymentUnpaid,
		FirstPayment:  firstPayment,
	}

	if err := u.BookingRepo.Create(tx, booking); err != nil {
		u.Log.WithError(err).Error("FAILED TO CREATE BOOKING.")
		return nil, fiber.ErrInternalServerError
	}

	// ── Update availability ───────────────────────────────────────────────────
	for _, date := range dates {
		dateStr := date.Format("2006-01-02")
		existing, exists := availMap[dateStr]
		if !exists {
			existing = rentable.Stock
		}

		newCount := existing - quantity
		if rentable.Type == "unit" {
			newCount = 0
		}

		avail := &entity.Availability{
			RentableID:     req.RentableID,
			Date:           date.Unix(),
			AvailableCount: newCount,
		}
		if err := u.AvailabilityRepo.Upsert(tx, avail); err != nil {
			u.Log.WithError(err).Error("FAILED TO UPDATE AVAILABILITY.")
			return nil, fiber.ErrInternalServerError
		}
	}

	if err := tx.Commit().Error; err != nil {
		u.Log.WithError(err).Error("FAILED TO COMMIT TRANSACTION.")
		return nil, fiber.ErrInternalServerError
	}

	// If caller provided a phone number and the user doesn't have one yet, persist it.
	if req.PhoneNumber != "" {
		user := new(entity.User)
		if err := u.UserRepo.FindById(u.DB, user, userID); err == nil && user.PhoneNumber == "" {
			user.PhoneNumber = req.PhoneNumber
			if err := u.UserRepo.Update(u.DB, user); err != nil {
				u.Log.WithError(err).Warn("failed to update user phone number")
			}
		}
	}

	u.notifyBookingCreated(booking)

	return model.BookingToResponse(booking), nil
}

func (u *BookingUseCase) GetByID(ctx context.Context, id string, userID string) (*model.BookingResponse, error) {
	tx := u.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	booking := new(entity.Booking)
	if err := u.BookingRepo.FindById(tx, booking, id); err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fiber.ErrNotFound
		}
		u.Log.WithError(err).Error("FAILED TO GET BOOKING.")
		return nil, fiber.ErrInternalServerError
	}

	if booking.UserID != userID {
		return nil, fiber.ErrForbidden
	}

	if err := tx.Commit().Error; err != nil {
		u.Log.WithError(err).Error("FAILED TO COMMIT TRANSACTION.")
		return nil, fiber.ErrInternalServerError
	}

	return model.BookingToResponse(booking), nil
}

func (u *BookingUseCase) GetMyBookings(ctx context.Context, userID string) ([]model.BookingResponse, error) {
	tx := u.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	var bookings []entity.Booking
	if err := u.BookingRepo.FindByUserID(tx, &bookings, userID); err != nil {
		u.Log.WithError(err).Error("FAILED TO GET BOOKINGS.")
		return nil, fiber.ErrInternalServerError
	}

	if err := tx.Commit().Error; err != nil {
		u.Log.WithError(err).Error("FAILED TO COMMIT TRANSACTION.")
		return nil, fiber.ErrInternalServerError
	}

	return model.BookingsToResponses(bookings), nil
}

// ConfirmBooking lets the host accept (WAITING_PAYMENT) or reject (UNAVAILABLE) a PENDING booking.
func (u *BookingUseCase) ConfirmBooking(ctx context.Context, bookingID, userEmail, newStatus string) (*model.BookingResponse, error) {
	tx := u.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	booking := new(entity.Booking)
	if err := u.BookingRepo.FindById(tx, booking, bookingID); err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fiber.ErrNotFound
		}
		return nil, fiber.ErrInternalServerError
	}

	if booking.Status != entity.StatusPending {
		return nil, fiber.NewError(fiber.StatusBadRequest, "booking bukan dalam status pending")
	}

	if err := u.checkHostOwns(tx, booking.PropertyID, userEmail); err != nil {
		return nil, err
	}

	booking.Status = entity.BookingStatus(newStatus)
	if err := u.BookingRepo.Update(tx, booking); err != nil {
		return nil, fiber.ErrInternalServerError
	}

	if err := tx.Commit().Error; err != nil {
		return nil, fiber.ErrInternalServerError
	}

	u.notifyBookingConfirmed(booking)

	return model.BookingToResponse(booking), nil
}

// Checkout lets the host advance a booking from CHECK_IN to REVIEW (guest has checked out).
func (u *BookingUseCase) Checkout(ctx context.Context, bookingID, userEmail string) (*model.BookingResponse, error) {
	tx := u.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	booking := new(entity.Booking)
	if err := u.BookingRepo.FindById(tx, booking, bookingID); err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fiber.ErrNotFound
		}
		return nil, fiber.ErrInternalServerError
	}

	if booking.Status != entity.StatusCheckIn {
		return nil, fiber.NewError(fiber.StatusBadRequest, "booking bukan dalam status check in")
	}

	if err := u.checkHostOwns(tx, booking.PropertyID, userEmail); err != nil {
		return nil, err
	}

	booking.Status = entity.StatusReview
	if err := u.BookingRepo.Update(tx, booking); err != nil {
		return nil, fiber.ErrInternalServerError
	}

	if err := tx.Commit().Error; err != nil {
		return nil, fiber.ErrInternalServerError
	}

	return model.BookingToResponse(booking), nil
}

// Complete lets the guest mark a REVIEW booking as DONE after leaving a review.
func (u *BookingUseCase) Complete(ctx context.Context, bookingID, userID string) (*model.BookingResponse, error) {
	tx := u.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	booking := new(entity.Booking)
	if err := u.BookingRepo.FindById(tx, booking, bookingID); err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fiber.ErrNotFound
		}
		return nil, fiber.ErrInternalServerError
	}

	if booking.UserID != userID {
		return nil, fiber.ErrForbidden
	}

	if booking.Status != entity.StatusReview {
		return nil, fiber.NewError(fiber.StatusBadRequest, "booking bukan dalam status review")
	}

	booking.Status = entity.StatusDone
	if err := u.BookingRepo.Update(tx, booking); err != nil {
		return nil, fiber.ErrInternalServerError
	}

	if err := tx.Commit().Error; err != nil {
		return nil, fiber.ErrInternalServerError
	}

	return model.BookingToResponse(booking), nil
}

// GetHostBookings returns all bookings for properties owned by the host.
func (u *BookingUseCase) GetHostBookings(ctx context.Context, userEmail string) ([]model.BookingResponse, error) {
	tx := u.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	host := new(entity.HostProfile)
	if err := u.HostProfileRepo.FindByEmail(tx, host, userEmail); err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fiber.NewError(fiber.StatusNotFound, "host profile tidak ditemukan")
		}
		return nil, fiber.ErrInternalServerError
	}

	var properties []entity.Property
	if err := u.PropertyRepo.FindByHostID(tx, &properties, host.ID); err != nil {
		return nil, fiber.ErrInternalServerError
	}

	if len(properties) == 0 {
		return []model.BookingResponse{}, nil
	}

	propertyIDs := make([]string, len(properties))
	for i, p := range properties {
		propertyIDs[i] = p.ID
	}

	var bookings []entity.Booking
	if err := u.BookingRepo.FindByPropertyIDs(tx, &bookings, propertyIDs); err != nil {
		return nil, fiber.ErrInternalServerError
	}

	if err := tx.Commit().Error; err != nil {
		return nil, fiber.ErrInternalServerError
	}

	return model.BookingsToResponses(bookings), nil
}

// checkHostOwns verifies the logged-in user (by email) is the owner of the property.
func (u *BookingUseCase) checkHostOwns(tx *gorm.DB, propertyID, userEmail string) error {
	property := new(entity.Property)
	if err := u.PropertyRepo.FindById(tx, property, propertyID); err != nil {
		return fiber.ErrInternalServerError
	}

	host := new(entity.HostProfile)
	if err := u.HostProfileRepo.FindById(tx, host, property.HostID); err != nil {
		return fiber.ErrInternalServerError
	}

	if host.Email != userEmail {
		return fiber.ErrForbidden
	}

	return nil
}

// ── Admin methods (no ownership check) ───────────────────────────────────────

func (u *BookingUseCase) GetAllBookings(ctx context.Context) ([]model.BookingResponse, error) {
	tx := u.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	var bookings []entity.Booking
	if err := u.BookingRepo.FindAll(tx, &bookings); err != nil {
		return nil, fiber.ErrInternalServerError
	}

	if err := tx.Commit().Error; err != nil {
		return nil, fiber.ErrInternalServerError
	}

	return model.BookingsToResponses(bookings), nil
}

func (u *BookingUseCase) AdminConfirmBooking(ctx context.Context, bookingID, newStatus string) (*model.BookingResponse, error) {
	tx := u.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	booking := new(entity.Booking)
	if err := u.BookingRepo.FindById(tx, booking, bookingID); err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fiber.ErrNotFound
		}
		return nil, fiber.ErrInternalServerError
	}

	if booking.Status != entity.StatusPending {
		return nil, fiber.NewError(fiber.StatusBadRequest, "booking bukan dalam status pending")
	}

	booking.Status = entity.BookingStatus(newStatus)
	if err := u.BookingRepo.Update(tx, booking); err != nil {
		return nil, fiber.ErrInternalServerError
	}

	if err := tx.Commit().Error; err != nil {
		return nil, fiber.ErrInternalServerError
	}

	u.notifyBookingConfirmed(booking)

	return model.BookingToResponse(booking), nil
}

func (u *BookingUseCase) AdminCheckout(ctx context.Context, bookingID string) (*model.BookingResponse, error) {
	tx := u.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	booking := new(entity.Booking)
	if err := u.BookingRepo.FindById(tx, booking, bookingID); err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fiber.ErrNotFound
		}
		return nil, fiber.ErrInternalServerError
	}

	if booking.Status != entity.StatusCheckIn {
		return nil, fiber.NewError(fiber.StatusBadRequest, "booking bukan dalam status check in")
	}

	booking.Status = entity.StatusReview
	if err := u.BookingRepo.Update(tx, booking); err != nil {
		return nil, fiber.ErrInternalServerError
	}

	if err := tx.Commit().Error; err != nil {
		return nil, fiber.ErrInternalServerError
	}

	return model.BookingToResponse(booking), nil
}

func (u *BookingUseCase) AdminComplete(ctx context.Context, bookingID string) (*model.BookingResponse, error) {
	tx := u.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	booking := new(entity.Booking)
	if err := u.BookingRepo.FindById(tx, booking, bookingID); err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fiber.ErrNotFound
		}
		return nil, fiber.ErrInternalServerError
	}

	if booking.Status != entity.StatusReview {
		return nil, fiber.NewError(fiber.StatusBadRequest, "booking bukan dalam status review")
	}

	booking.Status = entity.StatusDone
	if err := u.BookingRepo.Update(tx, booking); err != nil {
		return nil, fiber.ErrInternalServerError
	}

	if err := tx.Commit().Error; err != nil {
		return nil, fiber.ErrInternalServerError
	}

	return model.BookingToResponse(booking), nil
}
