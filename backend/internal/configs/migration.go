package configs

import (
	"log"

	"github.com/muhali16/listmak-service/internal/models"
	"gorm.io/gorm"
)

func AutoMigrate(db *gorm.DB) {
	// Cleanup legacy unique constraints created under the old `uniqueIndex` tag
	// (named idx_*). The models now use `unique`, so GORM manages constraints
	// named uni_*. Drop the old idx_-named ones so GORM recreates them cleanly
	// instead of trying (and failing) to DROP a constraint that never existed.
	if db.Migrator().HasTable("users") {
		db.Exec(`ALTER TABLE users DROP CONSTRAINT IF EXISTS idx_users_google_id`)
		db.Exec(`ALTER TABLE users DROP CONSTRAINT IF EXISTS idx_users_email`)
	}

	db.Exec(`CREATE EXTENSION IF NOT EXISTS pg_trgm`)
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_orders_detail_trgm ON orders USING GIN (order_detail gin_trgm_ops) WHERE deleted_at IS NULL`)

	if err := db.AutoMigrate(models.ModelRegistry()...); err != nil {
		log.Fatalf("AutoMigrate failed: %v", err)
	}
	log.Println("AutoMigrate completed")
}
