package http

import (
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
	"id.diengs.backend/internal/model"
	"id.diengs.backend/internal/usecase"
)

type RentableController struct {
	Log             *logrus.Logger
	RentableUseCase *usecase.RentableUseCase
}

func NewRentableController(log *logrus.Logger, rentableUseCase *usecase.RentableUseCase) *RentableController {
	return &RentableController{
		Log:             log,
		RentableUseCase: rentableUseCase,
	}
}

func (c *RentableController) Create(ctx *fiber.Ctx) error {
	req := new(model.RentableCreateRequest)
	if err := ctx.BodyParser(req); err != nil {
		c.Log.WithError(err).Error("failed to parse request body")
		return fiber.ErrBadRequest
	}

	response, err := c.RentableUseCase.Create(ctx.UserContext(), req)
	if err != nil {
		return err
	}

	if response == nil {
		c.Log.Error("failed to create rentable")
		return fiber.ErrInternalServerError
	}

	return ctx.JSON(model.WebResponse[model.RentableResponse]{
		Success: true,
		Message: "success create rentable",
		Data:    *response,
	})
}
