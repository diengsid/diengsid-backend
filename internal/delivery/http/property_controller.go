package http

import (
	"math"

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

func (c *PropertyController) Search(ctx *fiber.Ctx) error {
	req := &model.SearchPropertyRequest{
		Key:          ctx.Query("key"),
		CheckIn:      ctx.Query("check_in"),
		CheckOut:     ctx.Query("check_out"),
		GuestCount:   ctx.QueryInt("guest_count", 1),
		AttractionID: ctx.Query("attraction_id"),
		PropertyType: ctx.Query("property_type"),
		Page:         ctx.QueryInt("page", 1),
		Size:         ctx.QueryInt("size", 10),
	}

	responses, total, err := c.PropertyUseCase.Search(ctx.UserContext(), req)
	if err != nil {
		c.Log.WithError(err).Error("failed to search properties")
		return err
	}

	paging := &model.PageMetadata{
		Page:      req.Page,
		Size:      req.Size,
		TotalItem: total,
		TotalPage: int64(math.Ceil(float64(total) / float64(req.Size))),
	}

	return ctx.JSON(model.WebResponse[[]model.PropertyResponse]{
		Success: true,
		Message: "success search properties",
		Data:    responses,
		Paging:  paging,
	})
}

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

func (c *PropertyController) GetBySlug(ctx *fiber.Ctx) error {
	slug := ctx.Params("slug")

	response, err := c.PropertyUseCase.GetBySlug(ctx.UserContext(), slug)
	if err != nil {
		c.Log.WithError(err).Error("failed to get property by slug")
		return err
	}

	return ctx.JSON(model.WebResponse[model.PropertyResponse]{
		Success: true,
		Message: "success get property by slug",
		Data:    *response,
	})
}
