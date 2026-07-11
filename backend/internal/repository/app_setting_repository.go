package repository

import (
	"errors"

	"github.com/muhali16/listmak-service/internal/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type AppSettingRepository interface {
	// Get returns the value for key, or "" (no error) if unset.
	Get(key string) (string, error)
	Set(key, value string) error
}

type appSettingRepository struct {
	db *gorm.DB
}

func NewAppSettingRepository(db *gorm.DB) AppSettingRepository {
	return &appSettingRepository{db: db}
}

func (r *appSettingRepository) Get(key string) (string, error) {
	var s models.AppSetting
	err := r.db.Where("key = ?", key).First(&s).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return s.Value, nil
}

func (r *appSettingRepository) Set(key, value string) error {
	return r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "key"}},
		DoUpdates: clause.AssignmentColumns([]string{"value", "updated_at"}),
	}).Create(&models.AppSetting{Key: key, Value: value}).Error
}
