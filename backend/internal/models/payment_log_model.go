package models

import (
	"encoding/json"
	"time"
)

// PaymentLog is an append-only audit trail of every gateway interaction for a
// payment: transaction create, status detail checks, cancel, webhook, and
// completion. Stores the raw Pakasir response (safe — no secrets); the outbound
// request is NOT stored because it carries the api_key.
type PaymentLog struct {
	ID         uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	OrderID    string    `gorm:"type:varchar(40);index;not null" json:"order_id"`
	Action     string    `gorm:"type:varchar(20);index" json:"action"` // create|detail|cancel|webhook|complete
	StatusCode int       `json:"status_code"`                          // HTTP status from Pakasir (0 if n/a)
	Success    bool      `json:"success"`
	Response   json.RawMessage `gorm:"type:jsonb" json:"response"` // raw Pakasir response or webhook payload
	Error      string    `gorm:"type:text" json:"error,omitempty"`
	CreatedAt  time.Time `gorm:"<-:create;index" json:"created_at"`
}
