package usecase

import (
	"context"
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"id.diengs.backend/internal/entity"
	"id.diengs.backend/internal/model"
	"id.diengs.backend/internal/pkg"
	"id.diengs.backend/internal/repository"
)

type PaymentUseCase struct {
	DB          *gorm.DB
	Log         *logrus.Logger
	BookingRepo *repository.BookingRepo
	UserRepo    *repository.UserRepo
	PaymentRepo *repository.PaymentRepo
	Doku        *pkg.DokuClient
}

func NewPaymentUseCase(
	db *gorm.DB,
	log *logrus.Logger,
	bookingRepo *repository.BookingRepo,
	userRepo *repository.UserRepo,
	paymentRepo *repository.PaymentRepo,
	doku *pkg.DokuClient,
) *PaymentUseCase {
	return &PaymentUseCase{
		DB:          db,
		Log:         log,
		BookingRepo: bookingRepo,
		UserRepo:    userRepo,
		PaymentRepo: paymentRepo,
		Doku:        doku,
	}
}

// GetPaymentByBooking returns the existing payment record for a booking (owner only).
func (u *PaymentUseCase) GetPaymentByBooking(ctx context.Context, bookingID, userID string) (*model.PaymentInfoResponse, error) {
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

	invoiceNo := fmt.Sprintf("INV-%s", bookingID)
	payment := new(entity.Payment)
	if err := u.PaymentRepo.FindByInvoiceNo(tx, payment, invoiceNo); err != nil {
		// No payment record yet — return empty (not an error)
		return nil, nil
	}

	if err := tx.Commit().Error; err != nil {
		return nil, fiber.ErrInternalServerError
	}

	return &model.PaymentInfoResponse{
		InvoiceNo:  payment.InvoiceNo,
		Amount:     payment.Amount,
		Status:     string(payment.Status),
		PaymentURL: payment.PaymentURL,
	}, nil
}

// CreatePayment generates a DOKU payment URL for a WAITING_PAYMENT booking and saves the payment record.
func (u *PaymentUseCase) CreatePayment(ctx context.Context, bookingID, userID string) (*model.CreatePaymentResponse, error) {
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

	if booking.Status != entity.StatusWaiting {
		return nil, fiber.NewError(fiber.StatusBadRequest, "booking tidak dalam status menunggu pembayaran")
	}

	user := new(entity.User)
	if err := u.UserRepo.FindById(tx, user, userID); err != nil {
		return nil, fiber.ErrInternalServerError
	}

	invoiceNo := fmt.Sprintf("INV-%s", bookingID)
	amount := int64(booking.TotalPrice)
	description := fmt.Sprintf("Booking properti %d malam", booking.TotalNight)

	paymentURL, err := u.Doku.CreatePayment(invoiceNo, amount, description, pkg.DokuCustomer{
		Name:  user.Name,
		Email: user.Email,
	})
	if err != nil {
		u.Log.WithError(err).Error("failed to create doku payment")
		return nil, fiber.NewError(fiber.StatusBadGateway, "gagal membuat link pembayaran")
	}

	// Upsert: update URL if invoice already exists, otherwise create new record.
	existing := new(entity.Payment)
	if err := u.PaymentRepo.FindByInvoiceNo(tx, existing, invoiceNo); err == nil {
		existing.PaymentURL = paymentURL
		existing.Status = entity.PaymentRecordPending
		if err := u.PaymentRepo.Update(tx, existing); err != nil {
			u.Log.WithError(err).Error("failed to update payment record")
			return nil, fiber.ErrInternalServerError
		}
	} else {
		payment := &entity.Payment{
			BookingID:  bookingID,
			UserID:     userID,
			InvoiceNo:  invoiceNo,
			Amount:     amount,
			Status:     entity.PaymentRecordPending,
			PaymentURL: paymentURL,
		}
		if err := u.PaymentRepo.Create(tx, payment); err != nil {
			u.Log.WithError(err).Error("failed to save payment record")
			return nil, fiber.ErrInternalServerError
		}
	}

	if err := tx.Commit().Error; err != nil {
		return nil, fiber.ErrInternalServerError
	}

	return &model.CreatePaymentResponse{
		PaymentURL: paymentURL,
		InvoiceNo:  invoiceNo,
	}, nil
}

// HandleNotification processes DOKU's payment webhook and updates both the payment and booking records.
func (u *PaymentUseCase) HandleNotification(ctx context.Context, notif *model.DokuNotification) error {
	invoiceNo := notif.Order.InvoiceNumber
	bookingID := strings.TrimPrefix(invoiceNo, "INV-")

	tx := u.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	booking := new(entity.Booking)
	if err := u.BookingRepo.FindById(tx, booking, bookingID); err != nil {
		return fiber.ErrNotFound
	}

	payment := new(entity.Payment)
	_ = u.PaymentRepo.FindByInvoiceNo(tx, payment, invoiceNo)

	switch notif.Transaction.Status {
	case "SUCCESS":
		booking.PaymentStatus = entity.PaymentPaid
		booking.Status = entity.StatusCheckIn
		payment.Status = entity.PaymentRecordSuccess

	case "FAILED":
		booking.PaymentStatus = entity.PaymentUnpaid
		payment.Status = entity.PaymentRecordFailed

	case "EXPIRED":
		booking.PaymentStatus = entity.PaymentUnpaid
		payment.Status = entity.PaymentRecordExpired
	}

	if err := u.BookingRepo.Update(tx, booking); err != nil {
		u.Log.WithError(err).Error("failed to update booking after payment notification")
		return fiber.ErrInternalServerError
	}

	if payment.ID != "" {
		if err := u.PaymentRepo.Update(tx, payment); err != nil {
			u.Log.WithError(err).Error("failed to update payment record")
			return fiber.ErrInternalServerError
		}
	}

	return tx.Commit().Error
}
