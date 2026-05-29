package http

import (
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
	"id.diengs.backend/internal/model"
	"id.diengs.backend/internal/usecase"
)

type TouristAttractionController struct {
	Log     *logrus.Logger
	UseCase *usecase.TouristAttractionUseCase
}

func NewTouristAttractionController(log *logrus.Logger, uc *usecase.TouristAttractionUseCase) *TouristAttractionController {
	return &TouristAttractionController{Log: log, UseCase: uc}
}

// GET /api/tourist-attractions
func (c *TouristAttractionController) List(ctx *fiber.Ctx) error {
	resp, err := c.UseCase.List(ctx.Context())
	if err != nil {
		return err
	}
	return ctx.JSON(model.WebResponse[[]model.TouristAttractionResponse]{
		Success: true,
		Message: "success",
		Data:    resp,
	})
}

// POST /api/tourist-attractions
func (c *TouristAttractionController) Create(ctx *fiber.Ctx) error {
	req := new(model.TouristAttractionCreateRequest)
	if err := ctx.BodyParser(req); err != nil {
		return fiber.ErrBadRequest
	}
	resp, err := c.UseCase.Create(ctx.Context(), req)
	if err != nil {
		return err
	}
	return ctx.Status(fiber.StatusCreated).JSON(model.WebResponse[*model.TouristAttractionResponse]{
		Success: true,
		Message: "success create tourist attraction",
		Data:    resp,
	})
}

// GET /api/properties/:id/nearby-attractions
func (c *TouristAttractionController) GetNearbyByPropertyID(ctx *fiber.Ctx) error {
	propertyID := ctx.Params("id")
	resp, err := c.UseCase.GetNearbyByPropertyID(ctx.Context(), propertyID)
	if err != nil {
		return err
	}
	return ctx.JSON(model.WebResponse[[]model.NearbyAttractionResponse]{
		Success: true,
		Message: "success",
		Data:    resp,
	})
}

// PUT /api/properties/:id/nearby-attractions
func (c *TouristAttractionController) SetNearbyAttractions(ctx *fiber.Ctx) error {
	propertyID := ctx.Params("id")
	req := new(model.SetNearbyAttractionsRequest)
	if err := ctx.BodyParser(req); err != nil {
		return fiber.ErrBadRequest
	}
	resp, err := c.UseCase.SetNearbyAttractions(ctx.Context(), propertyID, req)
	if err != nil {
		return err
	}
	return ctx.JSON(model.WebResponse[[]model.NearbyAttractionResponse]{
		Success: true,
		Message: "success set nearby attractions",
		Data:    resp,
	})
}
