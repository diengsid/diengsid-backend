package usecase

import (
	"context"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"id.diengs.backend/internal/entity"
	"id.diengs.backend/internal/model"
	"id.diengs.backend/internal/repository"
)

type HostProfileUseCase struct {
	DB              *gorm.DB
	Log             *logrus.Logger
	Validate        *validator.Validate
	HostProfileRepo *repository.HostProfileRepo
}

func NewHostProfileUseCase(
	db *gorm.DB,
	log *logrus.Logger,
	validate *validator.Validate,
	hostProfileRepo *repository.HostProfileRepo,
) *HostProfileUseCase {
	return &HostProfileUseCase{
		DB:              db,
		Log:             log,
		Validate:        validate,
		HostProfileRepo: hostProfileRepo,
	}
}

// GET /api/hosts?key=
func (u *HostProfileUseCase) List(ctx context.Context, key string) ([]model.HostProfileResponse, error) {
	var hosts []entity.HostProfile
	if err := u.HostProfileRepo.FindAll(u.DB.WithContext(ctx), &hosts, key); err != nil {
		u.Log.WithError(err).Error("failed to list hosts")
		return nil, fiber.ErrInternalServerError
	}

	resp := make([]model.HostProfileResponse, 0, len(hosts))
	for _, h := range hosts {
		resp = append(resp, *model.HostToResponse(&h))
	}
	return resp, nil
}

// GET /api/hosts/:id
func (u *HostProfileUseCase) GetByID(ctx context.Context, id string) (*model.HostProfileResponse, error) {
	host := new(entity.HostProfile)
	if err := u.HostProfileRepo.FindById(u.DB.WithContext(ctx), host, id); err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fiber.ErrNotFound
		}
		u.Log.WithError(err).Error("failed to get host")
		return nil, fiber.ErrInternalServerError
	}
	return model.HostToResponse(host), nil
}

// POST /api/hosts
func (u *HostProfileUseCase) Create(ctx context.Context, req *model.HostCreateRequest) (*model.HostProfileResponse, error) {
	if err := u.Validate.Struct(req); err != nil {
		u.Log.WithError(err).Error("invalid host create request")
		return nil, fiber.ErrBadRequest
	}

	host := &entity.HostProfile{
		Name:              req.Name,
		Email:             req.Email,
		PhoneNumber:       req.PhoneNumber,
		ProfilePictureURL: req.ProfilePictureURL,
		Address:           req.Address,
		BankAccountName:   req.BankAccountName,
		BankAccountNumber: req.BankAccountNumber,
		KTPNumber:         req.KTPNumber,
		Bio:               req.Bio,
	}

	if err := u.HostProfileRepo.Create(u.DB.WithContext(ctx), host); err != nil {
		u.Log.WithError(err).Error("failed to create host")
		return nil, fiber.ErrInternalServerError
	}
	return model.HostToResponse(host), nil
}

// PUT /api/hosts/:id
func (u *HostProfileUseCase) Update(ctx context.Context, id string, req *model.HostUpdateRequest) (*model.HostProfileResponse, error) {
	db := u.DB.WithContext(ctx)

	host := new(entity.HostProfile)
	if err := u.HostProfileRepo.FindById(db, host, id); err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fiber.ErrNotFound
		}
		u.Log.WithError(err).Error("failed to find host for update")
		return nil, fiber.ErrInternalServerError
	}

	if req.Name != "" {
		host.Name = req.Name
	}
	if req.Email != "" {
		host.Email = req.Email
	}
	if req.PhoneNumber != "" {
		host.PhoneNumber = req.PhoneNumber
	}
	if req.ProfilePictureURL != "" {
		host.ProfilePictureURL = req.ProfilePictureURL
	}
	if req.Address != "" {
		host.Address = req.Address
	}
	if req.BankAccountName != "" {
		host.BankAccountName = req.BankAccountName
	}
	if req.BankAccountNumber != "" {
		host.BankAccountNumber = req.BankAccountNumber
	}
	if req.KTPNumber != "" {
		host.KTPNumber = req.KTPNumber
	}
	if req.Bio != "" {
		host.Bio = req.Bio
	}
	host.UpdatedAt = time.Now().UnixMilli()

	if err := u.HostProfileRepo.Update(db, host); err != nil {
		u.Log.WithError(err).Error("failed to update host")
		return nil, fiber.ErrInternalServerError
	}
	return model.HostToResponse(host), nil
}

// DELETE /api/hosts/:id
func (u *HostProfileUseCase) Delete(ctx context.Context, id string) error {
	db := u.DB.WithContext(ctx)

	host := new(entity.HostProfile)
	if err := u.HostProfileRepo.FindById(db, host, id); err != nil {
		if err == gorm.ErrRecordNotFound {
			return fiber.ErrNotFound
		}
		u.Log.WithError(err).Error("failed to find host for delete")
		return fiber.ErrInternalServerError
	}

	if err := u.HostProfileRepo.Delete(db, host); err != nil {
		u.Log.WithError(err).Error("failed to delete host")
		return fiber.ErrInternalServerError
	}
	return nil
}
