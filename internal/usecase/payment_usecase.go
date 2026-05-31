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
	"id.diengs.backend/internal/pkg/mailview"
	"id.diengs.backend/internal/pkg/message"
	"id.diengs.backend/internal/repository"
)

type PaymentUseCase struct {
	DB           *gorm.DB
	Log          *logrus.Logger
	BookingRepo  *repository.BookingRepo
	UserRepo     *repository.UserRepo
	PaymentRepo  *repository.PaymentRepo
	PropertyRepo *repository.PropertyRepo
	Doku         *pkg.DokuClient
	WA           pkg.WhatsAppSender
	Mail         *pkg.Mail
}

func NewPaymentUseCase(
	db *gorm.DB,
	log *logrus.Logger,
	bookingRepo *repository.BookingRepo,
	userRepo *repository.UserRepo,
	paymentRepo *repository.PaymentRepo,
	propertyRepo *repository.PropertyRepo,
	doku *pkg.DokuClient,
	wa pkg.WhatsAppSender,
	mail *pkg.Mail,
) *PaymentUseCase {
	return &PaymentUseCase{
		DB:           db,
		Log:          log,
		BookingRepo:  bookingRepo,
		UserRepo:     userRepo,
		PaymentRepo:  paymentRepo,
		PropertyRepo: propertyRepo,
		Doku:         doku,
		WA:           wa,
		Mail:         mail,
	}
}

func (u *PaymentUseCase) sendWA(phone, message string) {
	if phone == "" || u.WA == nil {
		return
	}
	go func() {
		if err := u.WA.SendOne(phone, message); err != nil {
			u.Log.WithError(err).Warn("failed to send whatsapp notification")
		}
	}()
}

func (u *PaymentUseCase) sendMail(to, subject, body string) {
	if to == "" || u.Mail == nil {
		return
	}
	go func() {
		if err := u.Mail.SendMail([]string{to}, subject, body); err != nil {
			u.Log.WithError(err).Warn("failed to send email notification")
		}
	}()
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

	u.sendWA(user.PhoneNumber, message.PaymentLinkCustomer(user.Name, bookingID, booking.TotalPrice, paymentURL))
	u.sendMail(user.Email, "Link Pembayaran Booking - Diengs.id",
		mailview.PaymentLinkMailView(user.Name, bookingID, paymentURL, booking.TotalPrice),
	)

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

	if err := tx.Commit().Error; err != nil {
		return err
	}

	user := new(entity.User)
	if err := u.UserRepo.FindById(u.DB, user, booking.UserID); err == nil {
		status := notif.Transaction.Status
		var msg, subject string
		switch status {
		case "SUCCESS":
			subject = "Pembayaran Berhasil - Diengs.id"
			msg = message.PaymentSuccessCustomer(user.Name, bookingID, booking.TotalPrice)
		case "FAILED":
			subject = "Pembayaran Gagal - Diengs.id"
			msg = message.PaymentFailedCustomer(user.Name, bookingID)
		case "EXPIRED":
			subject = "Link Pembayaran Kedaluwarsa - Diengs.id"
			msg = message.PaymentExpiredCustomer(user.Name, bookingID)
		}
		if msg != "" {
			u.sendWA(user.PhoneNumber, msg)
			u.sendMail(user.Email, subject,
				mailview.PaymentStatusMailView(user.Name, bookingID, status, booking.TotalPrice),
			)
		}
	}

	// Notify host when payment is successful
	if notif.Transaction.Status == "SUCCESS" {
		property := new(entity.Property)
		if err := u.PropertyRepo.FindById(u.DB, property, booking.PropertyID, "Host"); err == nil {
			checkIn := booking.CheckIn.Format("02 Jan 2006")
			checkOut := booking.CheckOut.Format("02 Jan 2006")

			user := new(entity.User)
			guestName := bookingID
			if err := u.UserRepo.FindById(u.DB, user, booking.UserID); err == nil {
				guestName = user.Name
			}

			u.sendWA(property.Host.PhoneNumber, message.PaymentSuccessHost(property.Host.Name, bookingID, guestName, property.Title, checkIn, checkOut, booking.GuestCount, booking.TotalPrice))
			u.sendMail(property.Host.Email, "Pembayaran Tamu Diterima - Diengs.id",
				mailview.BookingPaidHostMailView(property.Host.Name, bookingID, guestName, property.Title, checkIn, checkOut, booking.GuestCount, booking.TotalPrice),
			)
		}
	}

	return nil
}
