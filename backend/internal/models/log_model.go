package models

import "time"

type SystemLog struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	RequestID  string    `gorm:"index" json:"request_id"`
	Method     string    `json:"method"`
	Path       string    `json:"path"`
	StatusCode int       `json:"status_code"`
	Latency    string    `json:"latency"`
	ClientIP   string    `json:"client_ip"`
	ErrorMsg   string    `json:"error_msg"`
	CreatedAt  time.Time `json:"created_at"`
}
