package repository

import (
	"github.com/muhali16/listmak-service/internal/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type PriceCatalogRepository interface {
	GetAll() ([]models.PriceCatalog, error)
	FuzzyMatch(itemDetail string) (*models.PriceCatalog, error)
	UpsertBatch(entries []models.PriceCatalog) error
}

type priceCatalogRepository struct {
	db *gorm.DB
}

func NewPriceCatalogRepository(db *gorm.DB) PriceCatalogRepository {
	return &priceCatalogRepository{db: db}
}

func (r *priceCatalogRepository) GetAll() ([]models.PriceCatalog, error) {
	var entries []models.PriceCatalog
	err := r.db.Find(&entries).Error
	return entries, err
}

func (r *priceCatalogRepository) FuzzyMatch(itemDetail string) (*models.PriceCatalog, error) {
	var entry models.PriceCatalog
	likePattern := "%" + itemDetail + "%"
	err := r.db.Where("item_name ILIKE ? OR ? ILIKE CONCAT('%', item_name, '%')", likePattern, itemDetail).
		Order("LENGTH(item_name) DESC").
		First(&entry).Error
	if err != nil {
		return nil, err
	}
	return &entry, nil
}

func (r *priceCatalogRepository) UpsertBatch(entries []models.PriceCatalog) error {
	if len(entries) == 0 {
		return nil
	}
	return r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "vendor_name"}, {Name: "item_name"}},
		DoUpdates: clause.AssignmentColumns([]string{"price", "updated_at"}),
	}).Create(&entries).Error
}
