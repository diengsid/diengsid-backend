package repository

import (
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"id.diengs.backend/internal/entity"
)

type PaymentRepo struct {
	Repository[entity.Payment]
	Log *logrus.Logger
}

func NewPaymentRepo(log *logrus.Logger) *PaymentRepo {
	return &PaymentRepo{Log: log}
}

func (r *PaymentRepo) FindByInvoiceNo(db *gorm.DB, payment *entity.Payment, invoiceNo string) error {
	return db.Where("invoice_no = ?", invoiceNo).Take(payment).Error
}
