package repository

import (
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"id.diengs.backend/internal/entity"
)

type HostProfileRepo struct {
	Repository[entity.HostProfile]
	Log *logrus.Logger
}

func NewHostProfileRepo(log *logrus.Logger) *HostProfileRepo {
	return &HostProfileRepo{
		Log: log,
	}
}

func (r *HostProfileRepo) FindByEmail(db *gorm.DB, host *entity.HostProfile, email string) error {
	return db.Where("email = ?", email).Take(host).Error
}
