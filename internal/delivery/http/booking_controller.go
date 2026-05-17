package http

import (
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
	"id.diengs.backend/internal/model"
	"id.diengs.backend/internal/usecase"
)

type BookingController struct {
	Log            *logrus.Logger
	BookingUseCase *usecase.BookingUseCase
}

func NewBookingController(log *logrus.Logger, bookingUseCase *usecase.BookingUseCase) *BookingController {
	return &BookingController{
		Log:            log,
		BookingUseCase: bookingUseCase,
	}
}

func (c *BookingController) Create(ctx *fiber.Ctx) error {
	user := ctx.Locals("user").(*model.UserResponse)

	req := new(model.BookingCreateRequest)
	if err := ctx.BodyParser(req); err != nil {
		c.Log.WithError(err).Error("failed to parse request body")
		return fiber.ErrBadRequest
	}

	response, err := c.BookingUseCase.Create(ctx.UserContext(), user.ID, req)
	if err != nil {
		return err
	}

	return ctx.Status(fiber.StatusCreated).JSON(model.WebResponse[*model.BookingResponse]{
		Success: true,
		Message: "success create booking",
		Data:    response,
	})
}

func (c *BookingController) GetByID(ctx *fiber.Ctx) error {
	user := ctx.Locals("user").(*model.UserResponse)
	id := ctx.Params("id")

	response, err := c.BookingUseCase.GetByID(ctx.UserContext(), id, user.ID)
	if err != nil {
		return err
	}

	return ctx.JSON(model.WebResponse[*model.BookingResponse]{
		Success: true,
		Message: "success get booking",
		Data:    response,
	})
}

func (c *BookingController) GetMyBookings(ctx *fiber.Ctx) error {
	user := ctx.Locals("user").(*model.UserResponse)

	responses, err := c.BookingUseCase.GetMyBookings(ctx.UserContext(), user.ID)
	if err != nil {
		return err
	}

	return ctx.JSON(model.WebResponse[[]model.BookingResponse]{
		Success: true,
		Message: "success get my bookings",
		Data:    responses,
	})
}

func (c *BookingController) GetHostBookings(ctx *fiber.Ctx) error {
	user := ctx.Locals("user").(*model.UserResponse)

	responses, err := c.BookingUseCase.GetHostBookings(ctx.UserContext(), user.Email)
	if err != nil {
		return err
	}

	return ctx.JSON(model.WebResponse[[]model.BookingResponse]{
		Success: true,
		Message: "success get host bookings",
		Data:    responses,
	})
}

func (c *BookingController) ConfirmBooking(ctx *fiber.Ctx) error {
	user := ctx.Locals("user").(*model.UserResponse)
	id := ctx.Params("id")

	req := new(model.ConfirmBookingRequest)
	if err := ctx.BodyParser(req); err != nil {
		return fiber.ErrBadRequest
	}

	response, err := c.BookingUseCase.ConfirmBooking(ctx.UserContext(), id, user.Email, req.Status)
	if err != nil {
		return err
	}

	return ctx.JSON(model.WebResponse[*model.BookingResponse]{
		Success: true,
		Message: "success confirm booking",
		Data:    response,
	})
}

func (c *BookingController) Checkout(ctx *fiber.Ctx) error {
	user := ctx.Locals("user").(*model.UserResponse)
	id := ctx.Params("id")

	response, err := c.BookingUseCase.Checkout(ctx.UserContext(), id, user.Email)
	if err != nil {
		return err
	}

	return ctx.JSON(model.WebResponse[*model.BookingResponse]{
		Success: true,
		Message: "success checkout booking",
		Data:    response,
	})
}

func (c *BookingController) Complete(ctx *fiber.Ctx) error {
	user := ctx.Locals("user").(*model.UserResponse)
	id := ctx.Params("id")

	response, err := c.BookingUseCase.Complete(ctx.UserContext(), id, user.ID)
	if err != nil {
		return err
	}

	return ctx.JSON(model.WebResponse[*model.BookingResponse]{
		Success: true,
		Message: "success complete booking",
		Data:    response,
	})
}

// ── Admin handlers ─────────────────────────────────────────────────────────────

func (c *BookingController) GetAllBookings(ctx *fiber.Ctx) error {
	responses, err := c.BookingUseCase.GetAllBookings(ctx.UserContext())
	if err != nil {
		return err
	}

	return ctx.JSON(model.WebResponse[[]model.BookingResponse]{
		Success: true,
		Message: "success get all bookings",
		Data:    responses,
	})
}

func (c *BookingController) AdminConfirmBooking(ctx *fiber.Ctx) error {
	id := ctx.Params("id")

	req := new(model.ConfirmBookingRequest)
	if err := ctx.BodyParser(req); err != nil {
		return fiber.ErrBadRequest
	}

	response, err := c.BookingUseCase.AdminConfirmBooking(ctx.UserContext(), id, req.Status)
	if err != nil {
		return err
	}

	return ctx.JSON(model.WebResponse[*model.BookingResponse]{
		Success: true,
		Message: "success confirm booking",
		Data:    response,
	})
}

func (c *BookingController) AdminCheckout(ctx *fiber.Ctx) error {
	id := ctx.Params("id")

	response, err := c.BookingUseCase.AdminCheckout(ctx.UserContext(), id)
	if err != nil {
		return err
	}

	return ctx.JSON(model.WebResponse[*model.BookingResponse]{
		Success: true,
		Message: "success checkout booking",
		Data:    response,
	})
}

func (c *BookingController) AdminComplete(ctx *fiber.Ctx) error {
	id := ctx.Params("id")

	response, err := c.BookingUseCase.AdminComplete(ctx.UserContext(), id)
	if err != nil {
		return err
	}

	return ctx.JSON(model.WebResponse[*model.BookingResponse]{
		Success: true,
		Message: "success complete booking",
		Data:    response,
	})
}
