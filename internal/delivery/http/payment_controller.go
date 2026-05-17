package http

import (
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
	"id.diengs.backend/internal/delivery/middleware"
	"id.diengs.backend/internal/model"
	"id.diengs.backend/internal/pkg"
	"id.diengs.backend/internal/usecase"
)

type PaymentController struct {
	Log            *logrus.Logger
	PaymentUseCase *usecase.PaymentUseCase
	Doku           *pkg.DokuClient
}

func NewPaymentController(log *logrus.Logger, paymentUseCase *usecase.PaymentUseCase, doku *pkg.DokuClient) *PaymentController {
	return &PaymentController{
		Log:            log,
		PaymentUseCase: paymentUseCase,
		Doku:           doku,
	}
}

// GetPaymentByBooking handles GET /api/bookings/:id/payment
func (c *PaymentController) GetPaymentByBooking(ctx *fiber.Ctx) error {
	userID := middleware.GetUser(ctx).ID
	bookingID := ctx.Params("id")

	resp, err := c.PaymentUseCase.GetPaymentByBooking(ctx.Context(), bookingID, userID)
	if err != nil {
		return err
	}

	return ctx.JSON(model.WebResponse[*model.PaymentInfoResponse]{Data: resp})
}

// CreatePayment handles POST /api/bookings/:id/pay
func (c *PaymentController) CreatePayment(ctx *fiber.Ctx) error {
	userID := middleware.GetUser(ctx).ID
	bookingID := ctx.Params("id")

	resp, err := c.PaymentUseCase.CreatePayment(ctx.Context(), bookingID, userID)
	if err != nil {
		return err
	}

	return ctx.JSON(model.WebResponse[*model.CreatePaymentResponse]{Data: resp})
}

// HandleNotification handles POST /api/payment/notify (DOKU webhook, no auth)
func (c *PaymentController) HandleNotification(ctx *fiber.Ctx) error {
	requestID := ctx.Get("Request-Id")
	requestTimestamp := ctx.Get("Request-Timestamp")
	signatureHeader := ctx.Get("Signature")
	rawBody := string(ctx.Body())

	c.Log.WithFields(logrus.Fields{
		"request_id":        requestID,
		"request_timestamp": requestTimestamp,
		"signature":         signatureHeader,
		"body":              rawBody,
	}).Info("doku notification received")

	if !c.Doku.VerifyNotification(requestID, requestTimestamp, rawBody, signatureHeader) {
		c.Log.WithFields(logrus.Fields{
			"request_id": requestID,
			"signature":  signatureHeader,
		}).Warn("doku notification signature mismatch — check secret_key and notification path")
		return ctx.SendStatus(fiber.StatusUnauthorized)
	}

	notif := new(model.DokuNotification)
	if err := ctx.BodyParser(notif); err != nil {
		c.Log.WithError(err).Error("failed to parse doku notification body")
		return fiber.ErrBadRequest
	}

	c.Log.WithFields(logrus.Fields{
		"invoice":            notif.Order.InvoiceNumber,
		"transaction_status": notif.Transaction.Status,
		"amount":             notif.Order.Amount,
	}).Info("processing doku notification")

	if err := c.PaymentUseCase.HandleNotification(ctx.Context(), notif); err != nil {
		c.Log.WithError(err).Error("handle doku notification failed")
		return err
	}

	c.Log.WithField("invoice", notif.Order.InvoiceNumber).Info("doku notification processed successfully")
	return ctx.SendStatus(fiber.StatusOK)
}
