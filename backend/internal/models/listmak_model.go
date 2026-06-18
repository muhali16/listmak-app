package models

import (
	"time"

	"gorm.io/gorm"
)

type Listmak struct {
	ID          uint           `gorm:"primaryKey;autoIncrement" json:"id"`
	Title       string         `gorm:"type:varchar(255)" json:"title"`
	Date        time.Time      `gorm:"type:date;not null;index" json:"date"`
	CreatedBy   *uint          `json:"created_by"`
	User        *User          `gorm:"foreignKey:CreatedBy" json:"user,omitempty"`
	TotalOrders int            `gorm:"default:0" json:"total_orders"`
	TotalAmount float64        `gorm:"type:decimal(12,2);default:0" json:"total_amount"`
	PaidAmount  float64        `gorm:"type:decimal(12,2);default:0" json:"paid_amount"`
	Status      string         `gorm:"type:varchar(20);default:'active'" json:"status"` // 'active', 'completed', 'cancelled'
	CreatedAt   time.Time      `gorm:"<-:create" json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
	Orders      []Order        `gorm:"foreignKey:ListmakID;constraint:OnDelete:CASCADE" json:"orders,omitempty"`
	ShareLinks  []ShareLink    `gorm:"foreignKey:ListmakID" json:"share_links,omitempty"`
}
