package http

import (
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
	"id.diengs.backend/internal/model"
	"id.diengs.backend/internal/usecase"
)

type PropertyController struct {
	Log             *logrus.Logger
	PropertyUseCase *usecase.PropertyUseCase
}

func NewPropertyController(
	log *logrus.Logger,
	propertyUseCase *usecase.PropertyUseCase,
) *PropertyController {
	return &PropertyController{
		Log:             log,
		PropertyUseCase: propertyUseCase,
	}
}

// create property
func (c *PropertyController) Create(ctx *fiber.Ctx) error {
	req := new(model.PropertyCreateRequest)
	if err := ctx.BodyParser(req); err != nil {
		c.Log.WithError(err).Error("failed to parse request body")
		return fiber.ErrBadRequest
	}

	response, err := c.PropertyUseCase.Create(ctx.UserContext(), req)
	if err != nil {
		return err
	}

	return ctx.JSON(model.WebResponse[model.PropertyResponse]{
		Success: true,
		Message: "success create property",
		Data:    *response,
	})
}

// get property by id
func (c *PropertyController) GetByID(ctx *fiber.Ctx) error {
	id := ctx.Params("id")

	response, err := c.PropertyUseCase.GetByID(ctx.UserContext(), id)
	if err != nil {
		c.Log.WithError(err).Error("failed to get property by id")
		return err
	}

	return ctx.JSON(model.WebResponse[model.PropertyResponse]{
		Success: true,
		Message: "success get property by id",
		Data:    *response,
	})
}
