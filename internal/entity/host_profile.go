package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type HostProfile struct {
	ID                string `gorm:"column:id;primaryKey"`
	Name              string `gorm:"column:name"`
	Email             string `gorm:"column:email"`
	PhoneNumber       string `gorm:"column:phone_number"`
	ProfilePictureURL string `gorm:"column:profile_picture_url"`
	Address           string `gorm:"column:address"`
	BankAccountName   string `gorm:"column:bank_account_name"`
	BankAccountNumber string `gorm:"column:bank_account_number"`
	KTPNumber         string `gorm:"column:ktp_number"`
	Bio               string `gorm:"column:bio"`

	CreatedAt int64 `gorm:"column:created_at"`
	UpdatedAt int64 `gorm:"column:updated_at"`
}

func (HostProfile) TableName() string {
	return "host_profiles"
}

func (h *HostProfile) BeforeCreate(tx *gorm.DB) (err error) {
	h.ID = uuid.NewString()
	h.CreatedAt = time.Now().UnixMilli()
	h.UpdatedAt = time.Now().UnixMilli()
	return nil
}
