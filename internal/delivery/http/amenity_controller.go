package http

import (
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
	"id.diengs.backend/internal/model"
	"id.diengs.backend/internal/usecase"
)

type AmenityController struct {
	Log            *logrus.Logger
	AmenityUseCase *usecase.AmenityUseCase
}

func NewAmenityController(log *logrus.Logger, amenityUseCase *usecase.AmenityUseCase) *AmenityController {
	return &AmenityController{Log: log, AmenityUseCase: amenityUseCase}
}

// GET /api/amenities
func (c *AmenityController) List(ctx *fiber.Ctx) error {
	resp, err := c.AmenityUseCase.List(ctx.Context())
	if err != nil {
		return err
	}
	return ctx.JSON(model.WebResponse[[]model.AmenityResponse]{
		Success: true,
		Message: "success",
		Data:    resp,
	})
}

// POST /api/amenities
func (c *AmenityController) Create(ctx *fiber.Ctx) error {
	req := new(model.AmenityCreateRequest)
	if err := ctx.BodyParser(req); err != nil {
		return fiber.ErrBadRequest
	}
	resp, err := c.AmenityUseCase.Create(ctx.Context(), req)
	if err != nil {
		return err
	}
	return ctx.Status(fiber.StatusCreated).JSON(model.WebResponse[*model.AmenityResponse]{
		Success: true,
		Message: "success create amenity",
		Data:    resp,
	})
}

// PUT /api/properties/:id/amenities
func (c *AmenityController) SetPropertyAmenities(ctx *fiber.Ctx) error {
	propertyID := ctx.Params("id")
	req := new(model.SetAmenitiesRequest)
	if err := ctx.BodyParser(req); err != nil {
		return fiber.ErrBadRequest
	}
	resp, err := c.AmenityUseCase.SetPropertyAmenities(ctx.Context(), propertyID, req)
	if err != nil {
		return err
	}
	return ctx.JSON(model.WebResponse[[]model.AmenityResponse]{
		Success: true,
		Message: "success set property amenities",
		Data:    resp,
	})
}

// PUT /api/rentables/:id/amenities
func (c *AmenityController) SetRentableAmenities(ctx *fiber.Ctx) error {
	rentableID := ctx.Params("id")
	req := new(model.SetAmenitiesRequest)
	if err := ctx.BodyParser(req); err != nil {
		return fiber.ErrBadRequest
	}
	resp, err := c.AmenityUseCase.SetRentableAmenities(ctx.Context(), rentableID, req)
	if err != nil {
		return err
	}
	return ctx.JSON(model.WebResponse[[]model.AmenityResponse]{
		Success: true,
		Message: "success set rentable amenities",
		Data:    resp,
	})
}
