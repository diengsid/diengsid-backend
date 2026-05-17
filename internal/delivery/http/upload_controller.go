package http

import (
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
	"id.diengs.backend/internal/model"
	"id.diengs.backend/internal/usecase"
)

type UploadController struct {
	Log           *logrus.Logger
	UploadUseCase *usecase.UploadUseCase
}

func NewUploadController(log *logrus.Logger, uploadUseCase *usecase.UploadUseCase) *UploadController {
	return &UploadController{
		Log:           log,
		UploadUseCase: uploadUseCase,
	}
}

// upload one file
func (c *UploadController) Upload(ctx *fiber.Ctx) error {
	file, err := ctx.FormFile("file")
	if err != nil {
		c.Log.WithError(err).Error("Failed to get file from form data")
		return fiber.ErrBadRequest
	}

	request := &model.UploadRequest{
		File: file,
	}

	response, err := c.UploadUseCase.Upload(ctx.UserContext(), request)
	if err != nil {
		c.Log.WithError(err).Error("Failed to create file")
		return err
	}

	return ctx.JSON(model.WebResponse[model.UploadResponse]{
		Success: true,
		Message: "success create experience",
		Data:    *response,
	})
}

// upload many file
func (c *UploadController) Uploads(ctx *fiber.Ctx) error {
	form, err := ctx.MultipartForm()
	if err != nil {
		c.Log.WithError(err).Error("Failed to get multipart form")
		return fiber.ErrBadRequest
	}

	files := form.File["files"] // ✅ ini array

	if err != nil {
		c.Log.WithError(err).Error("Failed to get file from form data")
		return fiber.ErrBadRequest
	}

	request := &model.UploadsRequest{
		Files: files,
	}

	response, err := c.UploadUseCase.Uploads(ctx.UserContext(), request)
	if err != nil {
		c.Log.WithError(err).Error("Failed to create file")
		return err
	}

	return ctx.JSON(model.WebResponse[model.UploadResponses]{
		Success: true,
		Message: "success create experience",
		Data:    *response,
	})
}
