package models

import (
	"time"

	"gorm.io/gorm"
)

// Payment represents a single guest checkout against a listmak via Pakasir.
// One payment covers one guest submission (the set of Orders whose PaymentID
// points here). Money is stored as int64 rupiah (IDR has no sub-unit) to avoid
// float rounding on the money path.
type Payment struct {
	ID            uint   `gorm:"primaryKey;autoIncrement" json:"id"`
	OrderID       string `gorm:"type:varchar(40);unique;not null" json:"order_id"` // our invoice id == Pakasir order_id
	ListmakID     uint   `gorm:"not null;index" json:"listmak_id"`
	GuestName     string `gorm:"type:varchar(100);not null" json:"guest_name"`
	GuestWhatsapp string `gorm:"type:varchar(20);not null" json:"guest_whatsapp"`

	Amount        int64  `gorm:"not null" json:"amount"`         // base amount we sent to Pakasir (IDR)
	Fee           int64  `gorm:"default:0" json:"fee"`           // fee added by Pakasir
	TotalPayment  int64  `gorm:"default:0" json:"total_payment"` // amount+fee, what the customer actually pays
	PaymentMethod string `gorm:"type:varchar(30)" json:"payment_method"`
	PaymentNumber string `gorm:"type:text" json:"payment_number"` // QR string or Virtual Account number

	Status      string     `gorm:"type:varchar(20);default:'pending';index" json:"status"` // pending|completed|cancelled|expired|failed
	ExpiresAt   *time.Time `json:"expires_at"`
	CompletedAt *time.Time `json:"completed_at"`

	// IsSandbox marks a test-mode transaction (per Pakasir). Sandbox payments are
	// excluded from revenue stats and badged in admin.
	IsSandbox bool `gorm:"default:false;index" json:"is_sandbox"`

	// Fulfilled is true once the paid orders were materialized. A completed
	// payment with Fulfilled=false means money arrived but the listmak was already
	// closed — no orders created; needs a manual refund/review by admin.
	Fulfilled bool `gorm:"default:false" json:"fulfilled"`

	// ItemsSnapshot holds the guest's order lines as JSON. Orders are NOT written
	// to the shared list until payment completes; on completion they are created
	// from this snapshot. Keeps unpaid/abandoned checkouts out of the ledger.
	ItemsSnapshot []byte `gorm:"type:jsonb" json:"-"`

	CreatedAt time.Time      `gorm:"<-:create" json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// PaymentItem is one order line captured at checkout, stored in
// Payment.ItemsSnapshot and materialized into Orders once payment completes.
type PaymentItem struct {
	Name        string  `json:"name"`
	OrderDetail string  `json:"order_detail"`
	VendorName  string  `json:"vendor_name"`
	Price       float64 `json:"price"`
	Qty         int     `json:"qty"`
}
