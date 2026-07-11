package models

import (
	"time"

	"gorm.io/gorm"
)

type Order struct {
	ID          uint           `gorm:"primaryKey;autoIncrement" json:"id"`
	ListmakID   uint           `gorm:"not null;index" json:"listmak_id"`
	Name        string         `gorm:"type:varchar(100);not null;index" json:"name"`
	OrderDetail string         `gorm:"type:text;not null" json:"order_detail"`
	Price       float64        `gorm:"type:decimal(12,2);default:0" json:"price"`
	Qty         int            `gorm:"default:1" json:"qty"`
	TotalPrice  float64        `gorm:"type:decimal(12,2);generated:always as (price * qty) stored;<-:false" json:"total_price"` // Generated column supported in postgres/mysql 5.7+
	IsPaid      bool           `gorm:"default:false;index" json:"is_paid"`
	PaidAt      *time.Time     `json:"paid_at"`
	PaymentID   *uint          `gorm:"index" json:"payment_id,omitempty"` // set when this order is covered by a guest Payment
	PaidAmount  float64        `gorm:"type:decimal(12,2);default:0" json:"paid_amount"` // frozen amount paid via QR at payment time; unaffected by later price edits
	VendorName  string         `gorm:"type:varchar(100)" json:"vendor_name"`
	AddedVia    string         `gorm:"type:varchar(20);default:'parse'" json:"added_via"` // 'parse', 'manual', 'sharelink'
	AddedAt     time.Time      `gorm:"default:CURRENT_TIMESTAMP;<-:create" json:"added_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// Hook untuk update total di listmak bisa diimplementasikan di service atau via database trigger seperti di dokumentasi.
// Karena dokumentasi menyebutkan trigger database, kita asumsikan database menangani kalkulasi agregat atau kita handle di service agar sinkron.
// GORM hooks juga bisa digunakan.
