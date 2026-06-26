package repository

import (
	"github.com/muhali16/listmak-service/internal/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type SummaryRepository interface {
	GetByListmakID(listmakID uint) (*models.ListmakSummary, error)
	Upsert(summary *models.ListmakSummary) error
	DeleteByListmakID(listmakID uint) error
}

type summaryRepository struct {
	db *gorm.DB
}

func NewSummaryRepository(db *gorm.DB) SummaryRepository {
	return &summaryRepository{db: db}
}

func (r *summaryRepository) GetByListmakID(listmakID uint) (*models.ListmakSummary, error) {
	var s models.ListmakSummary
	err := r.db.Where("listmak_id = ?", listmakID).First(&s).Error
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *summaryRepository) DeleteByListmakID(listmakID uint) error {
	return r.db.Where("listmak_id = ?", listmakID).Delete(&models.ListmakSummary{}).Error
}

func (r *summaryRepository) Upsert(summary *models.ListmakSummary) error {
	return r.db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "listmak_id"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"summary_data", "order_watermark", "total_estimated",
			"total_actual", "generated_at", "confirmed_at",
		}),
	}).Create(summary).Error
}
