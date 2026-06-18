package models

import (
	"time"
)

type ShareLink struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	ShareID   string    `gorm:"type:varchar(20);unique;not null" json:"share_id"`
	ListmakID uint      `gorm:"not null;index" json:"listmak_id"`
	Listmak   Listmak   `gorm:"foreignKey:ListmakID" json:"listmak,omitempty"`
	Title     string    `gorm:"type:varchar(255)" json:"title"`
	ExpiresAt time.Time `gorm:"not null;index" json:"expires_at"`
	IsActive  bool      `gorm:"default:true" json:"is_active"`
	CreatedBy *uint     `json:"created_by"`
	CreatedAt time.Time `gorm:"<-:create" json:"created_at"`
}
