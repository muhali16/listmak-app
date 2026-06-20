package repository

import (
	"time"

	"github.com/muhali16/listmak-service/internal/models"
	"gorm.io/gorm"
)

type SystemLogFilter struct {
	RequestID  string
	Method     string
	StatusCode int
	From       *time.Time
	To         *time.Time
}

type SystemLogRepository interface {
	Create(log *models.SystemLog) error
	GetAll(page, limit int, f SystemLogFilter) ([]models.SystemLog, int64, error)
}

type systemLogRepository struct {
	db *gorm.DB
}

func NewSystemLogRepository(db *gorm.DB) SystemLogRepository {
	return &systemLogRepository{db: db}
}

func (r *systemLogRepository) Create(log *models.SystemLog) error {
	return r.db.Create(log).Error
}

func (r *systemLogRepository) GetAll(page, limit int, f SystemLogFilter) ([]models.SystemLog, int64, error) {
	var logs []models.SystemLog
	var total int64
	offset := (page - 1) * limit

	q := r.db.Model(&models.SystemLog{})
	if f.RequestID != "" {
		q = q.Where("request_id LIKE ?", "%"+f.RequestID+"%")
	}
	if f.Method != "" {
		q = q.Where("method = ?", f.Method)
	}
	if f.StatusCode > 0 {
		q = q.Where("status_code >= ? AND status_code < ?", f.StatusCode, f.StatusCode+100)
	}
	if f.From != nil {
		q = q.Where("created_at >= ?", f.From)
	}
	if f.To != nil {
		q = q.Where("created_at <= ?", f.To)
	}

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := q.Order("created_at desc").Offset(offset).Limit(limit).Find(&logs).Error; err != nil {
		return nil, 0, err
	}
	return logs, total, nil
}
