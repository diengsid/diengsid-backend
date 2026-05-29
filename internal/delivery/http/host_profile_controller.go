package http

import (
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
	"id.diengs.backend/internal/model"
	"id.diengs.backend/internal/usecase"
)

type HostProfileController struct {
	Log                *logrus.Logger
	HostProfileUseCase *usecase.HostProfileUseCase
}

func NewHostProfileController(log *logrus.Logger, uc *usecase.HostProfileUseCase) *HostProfileController {
	return &HostProfileController{Log: log, HostProfileUseCase: uc}
}

// GET /api/hosts?key=
func (c *HostProfileController) List(ctx *fiber.Ctx) error {
	key := ctx.Query("key")
	resp, err := c.HostProfileUseCase.List(ctx.Context(), key)
	if err != nil {
		return err
	}
	return ctx.JSON(model.WebResponse[[]model.HostProfileResponse]{
		Success: true,
		Message: "success",
		Data:    resp,
	})
}

// GET /api/hosts/:id
func (c *HostProfileController) GetByID(ctx *fiber.Ctx) error {
	id := ctx.Params("id")
	resp, err := c.HostProfileUseCase.GetByID(ctx.Context(), id)
	if err != nil {
		return err
	}
	return ctx.JSON(model.WebResponse[*model.HostProfileResponse]{
		Success: true,
		Message: "success",
		Data:    resp,
	})
}

// POST /api/hosts
func (c *HostProfileController) Create(ctx *fiber.Ctx) error {
	req := new(model.HostCreateRequest)
	if err := ctx.BodyParser(req); err != nil {
		return fiber.ErrBadRequest
	}
	resp, err := c.HostProfileUseCase.Create(ctx.Context(), req)
	if err != nil {
		return err
	}
	return ctx.Status(fiber.StatusCreated).JSON(model.WebResponse[*model.HostProfileResponse]{
		Success: true,
		Message: "success create host",
		Data:    resp,
	})
}

// PUT /api/hosts/:id
func (c *HostProfileController) Update(ctx *fiber.Ctx) error {
	id := ctx.Params("id")
	req := new(model.HostUpdateRequest)
	if err := ctx.BodyParser(req); err != nil {
		return fiber.ErrBadRequest
	}
	resp, err := c.HostProfileUseCase.Update(ctx.Context(), id, req)
	if err != nil {
		return err
	}
	return ctx.JSON(model.WebResponse[*model.HostProfileResponse]{
		Success: true,
		Message: "success update host",
		Data:    resp,
	})
}

// DELETE /api/hosts/:id
func (c *HostProfileController) Delete(ctx *fiber.Ctx) error {
	id := ctx.Params("id")
	if err := c.HostProfileUseCase.Delete(ctx.Context(), id); err != nil {
		return err
	}
	return ctx.JSON(model.WebResponse[any]{
		Success: true,
		Message: "success delete host",
	})
}
