package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	ID            string  `gorm:"column:id;primaryKey"`
	Name          string  `gorm:"column:name;not null"`
	Email         string  `gorm:"column:email;uniqueIndex;not null"`
	PhoneNumber   string  `gorm:"column:phone_number;default:''"`
	EmailVerified bool    `gorm:"column:email_verified;default:false"`
	Picture       *string `gorm:"column:picture"`
	Provider      *string `gorm:"provider"`
	ProviderID    *string `gorm:"provider_id"`
	Role          string  `gorm:"column:role;not null"`
	Password      string  `gorm:"column:password;not null"`

	Sessions []Session `gorm:"constraint:OnDelete:CASCADE"`

	CreatedAt int64 `gorm:"column:created_at"`
	UpdatedAt int64 `gorm:"column:updated_at"`
}

func (User) TableName() string {
	return "users"
}

func (u *User) BeforeCreate(tx *gorm.DB) (err error) {
	u.ID = uuid.NewString()
	u.CreatedAt = int64(time.Now().UnixMilli())
	u.UpdatedAt = int64(time.Now().UnixMilli())
	return nil
}

type Session struct {
	ID        string  `gorm:"column:id;primaryKey"`
	ExpiredAt int64   `gorm:"column:expired_at"`
	Token     string  `gorm:"column:token;uniqueIndex"`
	IPAddress *string `gorm:"column:ip_address"`
	UserAgent *string `gorm:"column:user_agent"`
	UserID    string  `gorm:"column:user_id;type:uuid;not null"`

	User User `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`

	CreatedAt int64 `gorm:"column:created_at"`
	UpdatedAt int64 `gorm:"column:updated_at"`
}

func (Session) TableName() string {
	return "sessions"
}

func (u *Session) BeforeCreate(tx *gorm.DB) (err error) {
	u.ID = uuid.NewString()
	u.CreatedAt = int64(time.Now().UnixMilli())
	u.UpdatedAt = int64(time.Now().UnixMilli())
	return nil
}

type EmailOtp struct {
	ID           string `gorm:"column:id;primaryKey"`
	Email        string `gorm:"column:email;not null"`
	OtpCode      string `gorm:"column:otp_code;not null"`
	ExpiredAt    int64  `gorm:"column:expired_at;not null"`
	IsUsed       bool   `gorm:"column:is_used;not null"`
	AttemptCount int    `gorm:"column:attempt_count;not null"`
	MaxAttempt   int    `gorm:"column:max_attempt;default:5;not null"`

	CreatedAt int64 `gorm:"column:created_at"`
	UpdatedAt int64 `gorm:"column:updated_at"`
}

func (EmailOtp) TableName() string {
	return "email_otps"
}

func (u *EmailOtp) BeforeCreate(tx *gorm.DB) (err error) {
	u.ID = uuid.NewString()
	u.CreatedAt = int64(time.Now().UnixMilli())
	u.UpdatedAt = int64(time.Now().UnixMilli())
	return nil
}
