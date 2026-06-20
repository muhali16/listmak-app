package repository

import (
	"time"

	"github.com/muhali16/listmak-service/internal/models"
	"gorm.io/gorm"
)

type ListmakRepository interface {
	GetAllListmaks(page, limit int, status string, startDate, endDate *time.Time, userId uint) ([]models.Listmak, int64, error)
	GetListmakById(id uint) (models.Listmak, error)
	GetListmakByDate(date time.Time) ([]models.Listmak, error)
	CreateListmak(listmak models.Listmak) (models.Listmak, error)
	UpdateListmak(listmak models.Listmak) (models.Listmak, error)
	DeleteListmak(id uint) error
}

type listmakRepository struct {
	db *gorm.DB
}

func NewListmakRepository(db *gorm.DB) ListmakRepository {
	return &listmakRepository{db: db}
}

func (r *listmakRepository) GetAllListmaks(page, limit int, status string, startDate, endDate *time.Time, userId uint) ([]models.Listmak, int64, error) {
	var listmaks []models.Listmak
	var total int64

	query := r.db.Model(&models.Listmak{}).Where("created_by = ?", userId)

	if status != "" {
		query = query.Where("status = ?", status)
	}
	if startDate != nil {
		query = query.Where("date >= ?", startDate)
	}
	if endDate != nil {
		query = query.Where("date <= ?", endDate)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	if err := query.Preload("Orders").Preload("ShareLinks", "is_active = ? AND expires_at > ?", true, time.Now()).Offset(offset).Limit(limit).Order("date desc").Find(&listmaks).Error; err != nil {
		return nil, 0, err
	}

	return listmaks, total, nil
}

func (r *listmakRepository) GetListmakById(id uint) (models.Listmak, error) {
	var listmak models.Listmak
	// Preload Orders and Active ShareLinks
	if err := r.db.Preload("Orders").Preload("User").Preload("ShareLinks", "is_active = ? AND expires_at > ?", true, time.Now()).First(&listmak, id).Error; err != nil {
		return models.Listmak{}, err
	}
	return listmak, nil
}

func (r *listmakRepository) GetListmakByDate(date time.Time) ([]models.Listmak, error) {
	var listmaks []models.Listmak
	// Format to 2006-01-02 to match date only
	dateStr := date.Format("2006-01-02")
	if err := r.db.Preload("Orders").Preload("User").Preload("ShareLinks", "is_active = ? AND expires_at > ?", true, time.Now()).Where("DATE(date) = ?", dateStr).Find(&listmaks).Error; err != nil {
		return nil, err
	}
	return listmaks, nil
}

func (r *listmakRepository) CreateListmak(listmak models.Listmak) (models.Listmak, error) {
	if err := r.db.Create(&listmak).Error; err != nil {
		return models.Listmak{}, err
	}
	return listmak, nil
}

func (r *listmakRepository) UpdateListmak(listmak models.Listmak) (models.Listmak, error) {
	if err := r.db.Save(&listmak).Error; err != nil {
		return models.Listmak{}, err
	}
	return listmak, nil
}

func (r *listmakRepository) DeleteListmak(id uint) error {
	return r.db.Delete(&models.Listmak{}, id).Error
}
