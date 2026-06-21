package models

import "time"

type AILog struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	RequestID string    `gorm:"type:varchar(36);index" json:"request_id"`
	OrderID   *uint     `gorm:"index" json:"order_id"`
	Input     string    `gorm:"type:text" json:"input"`
	Output    string    `gorm:"type:text" json:"output"`
	Model     string    `json:"model"`
	Provider  string    `json:"provider"`
	LatencyMs int64     `json:"latency_ms"`
	Status    string    `json:"status"`
	ErrorMsg  string    `json:"error_msg"`
	CreatedAt time.Time `json:"created_at"`
}
