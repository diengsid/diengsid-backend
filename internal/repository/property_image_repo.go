package repository

import (
	"github.com/sirupsen/logrus"
	"id.diengs.backend/internal/entity"
)

type PropertyImageRepo struct {
	Repository[entity.PropertyImage]
	Log *logrus.Logger
}

func NewPropertyImageRepo(log *logrus.Logger) *PropertyImageRepo {
	return &PropertyImageRepo{
		Log: log,
	}
}
