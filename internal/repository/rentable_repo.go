package repository

import (
	"github.com/sirupsen/logrus"
	"id.diengs.backend/internal/entity"
)

type RentableRepo struct {
	Repository[entity.Rentable]
	Log *logrus.Logger
}

func NewRentableRepo(log *logrus.Logger) *RentableRepo {
	return &RentableRepo{
		Log: log,
	}
}
