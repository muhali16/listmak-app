package repository

import (
	"github.com/muhali16/listmak-service/internal/models"
	"gorm.io/gorm"
)

type PaymentLogRepository interface {
	Create(log *models.PaymentLog) error
	ListByOrderID(orderID string) ([]models.PaymentLog, error)
}

type paymentLogRepository struct {
	db *gorm.DB
}

func NewPaymentLogRepository(db *gorm.DB) PaymentLogRepository {
	return &paymentLogRepository{db: db}
}

func (r *paymentLogRepository) Create(log *models.PaymentLog) error {
	return r.db.Create(log).Error
}

func (r *paymentLogRepository) ListByOrderID(orderID string) ([]models.PaymentLog, error) {
	var logs []models.PaymentLog
	err := r.db.Where("order_id = ?", orderID).Order("created_at asc").Find(&logs).Error
	return logs, err
}
