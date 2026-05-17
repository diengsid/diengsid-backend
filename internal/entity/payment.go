package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PaymentRecordStatus string

const (
	PaymentRecordPending  PaymentRecordStatus = "PENDING"
	PaymentRecordSuccess  PaymentRecordStatus = "SUCCESS"
	PaymentRecordFailed   PaymentRecordStatus = "FAILED"
	PaymentRecordExpired  PaymentRecordStatus = "EXPIRED"
)

type Payment struct {
	ID         string              `gorm:"column:id;primaryKey"`
	BookingID  string              `gorm:"column:booking_id;not null"`
	UserID     string              `gorm:"column:user_id;not null"`
	InvoiceNo  string              `gorm:"column:invoice_no;not null;uniqueIndex"`
	Amount     int64               `gorm:"column:amount;not null"`
	Status     PaymentRecordStatus `gorm:"column:status;not null;default:PENDING"`
	PaymentURL string              `gorm:"column:payment_url"`
	CreatedAt  int64               `gorm:"column:created_at"`
	UpdatedAt  int64               `gorm:"column:updated_at"`
}

func (Payment) TableName() string {
	return "payments"
}

func (p *Payment) BeforeCreate(tx *gorm.DB) error {
	p.ID = uuid.NewString()
	p.CreatedAt = time.Now().UnixMilli()
	p.UpdatedAt = time.Now().UnixMilli()
	return nil
}

func (p *Payment) BeforeUpdate(tx *gorm.DB) error {
	p.UpdatedAt = time.Now().UnixMilli()
	return nil
}
