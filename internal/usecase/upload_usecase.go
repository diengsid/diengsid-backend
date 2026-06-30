package usecase

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"id.diengs.backend/internal/model"
)

type UploadUseCase struct {
	Log       *logrus.Logger
	Validator *validator.Validate
	Viper     *viper.Viper
}

func NewUploadUseCase(log *logrus.Logger, validator *validator.Validate, viper *viper.Viper) *UploadUseCase {
	return &UploadUseCase{
		Log:       log,
		Validator: validator,
		Viper:     viper,
	}
}

// create one file
func (u *UploadUseCase) Upload(ctx context.Context, request *model.UploadRequest) (*model.UploadResponse, error) {
	if err := u.Validator.Struct(request); err != nil {
		u.Log.WithError(err).Error("Failed to validate request body")
		return nil, fiber.ErrBadRequest
	}

	// validasi ukuran (2MB)
	if request.File.Size > 10*1024*1024 {
		u.Log.Error("FAILED TO COMMIT TRANSACTION.")
		return nil, fiber.ErrBadRequest
	}

	baseUrl := u.Viper.GetString("app.url")
	url, err := u.Save(request.File, baseUrl)
	if err != nil {
		u.Log.Error("FAILED TO SAVE FILE.")
		return nil, fiber.ErrBadRequest
	}

	return &model.UploadResponse{
		Url: *url,
	}, nil
}

// create many file
func (u *UploadUseCase) Uploads(ctx context.Context, request *model.UploadsRequest) (*model.UploadResponses, error) {
	if err := u.Validator.Struct(request); err != nil {
		u.Log.WithError(err).Error("Failed to validate request body")
		return nil, fiber.ErrBadRequest
	}

	baseUrl := u.Viper.GetString("app.url")

	var results []string

	for _, file := range request.Files {

		// validasi size per file
		if file.Size > 10*1024*1024 {
			u.Log.Warn("skip file: too large")
			continue
		}

		url, err := u.Save(file, baseUrl)
		if err != nil {
			u.Log.Warn("skip file: failed to save")
			continue
		}

		results = append(results, *url)
	}

	if len(results) == 0 {
		return nil, fiber.ErrBadRequest
	}

	return &model.UploadResponses{Urls: results}, nil
}

func (s *UploadUseCase) Save(file *multipart.FileHeader, baseURL string) (*string, error) {
	// isi sama seperti logic di atas

	ext := filepath.Ext(file.Filename)

	// folder: uploads/YYYY/MM
	folder := time.Now().Format("2006/01")
	dir := fmt.Sprintf("./uploads/%s", folder)

	// pastikan folder ada
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return nil, err
	}

	filename := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)
	fullPath := fmt.Sprintf("%s/%s", dir, filename)

	// buka file dari request
	src, err := file.Open()
	if err != nil {
		return nil, err
	}
	defer src.Close()

	// buat file tujuan
	dst, err := os.Create(fullPath)
	if err != nil {
		return nil, err
	}
	defer dst.Close()

	// copy isi file
	if _, err := io.Copy(dst, src); err != nil {
		return nil, err
	}

	// return URL (pakai ENV biar clean)
	url := fmt.Sprintf("%s/uploads/%s/%s", baseURL, folder, filename)

	return &url, nil
}
