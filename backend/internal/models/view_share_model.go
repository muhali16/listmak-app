package models

import (
	"encoding/json"
	"time"
)

type ViewShare struct {
	ID           uint            `gorm:"primaryKey;autoIncrement" json:"id"`
	ViewID       string          `gorm:"type:varchar(20);unique;not null" json:"view_id"`
	ListmakID    uint            `gorm:"not null;index" json:"listmak_id"`
	Listmak      Listmak         `gorm:"foreignKey:ListmakID" json:"listmak,omitempty"`
	Title        string          `gorm:"type:varchar(255)" json:"title"`
	SnapshotData json.RawMessage `gorm:"type:jsonb" json:"snapshot_data" swaggertype:"object"` // Menggunakan jsonb untuk postgres atau json untuk mysql
	// IsLive distinguishes links created after the live-update change from legacy
	// links. Default false preserves snapshot behaviour for all pre-existing rows
	// (non-retroactive); links created from now on set it true to serve live data.
	IsLive    bool  `gorm:"default:false" json:"is_live"`
	CreatedBy *uint `json:"created_by"`
	CreatedAt    time.Time       `gorm:"<-:create" json:"created_at"`
}
