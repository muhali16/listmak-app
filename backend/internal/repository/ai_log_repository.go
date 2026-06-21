package repository

import (
	"time"

	"github.com/muhali16/listmak-service/internal/models"
	"gorm.io/gorm"
)

type AILogRepository interface {
	Create(log *models.AILog) error
	GetAll(page, limit int, status, search string) ([]models.AILog, int64, error)
	DeleteOlderThan(before time.Time) (int64, error)
}

type aiLogRepository struct {
	db *gorm.DB
}

func NewAILogRepository(db *gorm.DB) AILogRepository {
	return &aiLogRepository{db: db}
}

func (r *aiLogRepository) Create(log *models.AILog) error {
	return r.db.Create(log).Error
}

func (r *aiLogRepository) GetAll(page, limit int, status, search string) ([]models.AILog, int64, error) {
	var logs []models.AILog
	var total int64
	offset := (page - 1) * limit
	q := r.db.Model(&models.AILog{})
	if status != "" {
		q = q.Where("status = ?", status)
	}
	if search != "" {
		q = q.Where("input ILIKE ?", "%"+search+"%")
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := q.Order("created_at desc").Offset(offset).Limit(limit).Find(&logs).Error; err != nil {
		return nil, 0, err
	}
	return logs, total, nil
}

func (r *aiLogRepository) DeleteOlderThan(before time.Time) (int64, error) {
	result := r.db.Where("created_at < ?", before).Delete(&models.AILog{})
	return result.RowsAffected, result.Error
}
