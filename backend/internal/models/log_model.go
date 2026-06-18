package models

import (
	"log"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type SystemLog struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	RequestID  string    `gorm:"index" json:"request_id"` // ID Unik per request
	Method     string    `json:"method"`
	Path       string    `json:"path"`
	StatusCode int       `json:"status_code"`
	Latency    string    `json:"latency"`
	ClientIP   string    `json:"client_ip"`
	ErrorMsg   string    `json:"error_msg"`
	CreatedAt  time.Time `json:"created_at"`
}

var DBLog *gorm.DB

func InitLogDB() {
	var err error
	DBLog, err = gorm.Open(sqlite.Open("logs.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	log.Println("Database connection established")
	DBLog.AutoMigrate(&SystemLog{})
}
