package configs

import (
	"github.com/muhali16/listmak-service/internal/models"
	"gorm.io/gorm"
)

func AutoMigrate(db *gorm.DB) {
	db.AutoMigrate(
		models.ModelRegistry()...,
	)
}
