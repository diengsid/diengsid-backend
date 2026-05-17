package http

import (
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
	"id.diengs.backend/internal/model"
	"id.diengs.backend/internal/usecase"
)

type AvailabilityController struct {
	Log                 *logrus.Logger
	AvailabilityUseCase *usecase.AvailabilityUseCase
}

func NewAvailabilityController(log *logrus.Logger, availabilityUseCase *usecase.AvailabilityUseCase) *AvailabilityController {
	return &AvailabilityController{
		Log:                 log,
		AvailabilityUseCase: availabilityUseCase,
	}
}

func (c *AvailabilityController) Check(ctx *fiber.Ctx) error {
	rentableID := ctx.Params("id")

	req := new(model.CheckAvailabilityRequest)
	if err := ctx.QueryParser(req); err != nil {
		c.Log.WithError(err).Error("failed to parse query params")
		return fiber.ErrBadRequest
	}

	responses, err := c.AvailabilityUseCase.Check(ctx.UserContext(), rentableID, req)
	if err != nil {
		return err
	}

	return ctx.JSON(model.WebResponse[[]model.AvailabilityResponse]{
		Success: true,
		Message: "success check availability",
		Data:    responses,
	})
}
