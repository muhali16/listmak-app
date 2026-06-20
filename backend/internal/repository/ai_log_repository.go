package repository

import (
	"github.com/muhali16/listmak-service/internal/models"
	"gorm.io/gorm"
)

type AILogRepository interface {
	Create(log *models.AILog) error
	GetAll(page, limit int) ([]models.AILog, int64)
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

func (r *aiLogRepository) GetAll(page, limit int) ([]models.AILog, int64) {
	var logs []models.AILog
	var total int64
	offset := (page - 1) * limit
	r.db.Model(&models.AILog{}).Count(&total)
	r.db.Order("created_at desc").Offset(offset).Limit(limit).Find(&logs)
	return logs, total
}
