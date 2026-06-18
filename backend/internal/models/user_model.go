package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID        uint           `gorm:"primaryKey;autoIncrement" json:"id"`
	GoogleID  string         `gorm:"type:varchar(255);unique;not null" json:"google_id"`
	Email     string         `gorm:"type:varchar(255);unique;not null" json:"email"`
	Name      string         `gorm:"type:varchar(255);not null" json:"name"`
	Avatar    string         `gorm:"type:text" json:"avatar"`
	Role      string         `gorm:"type:varchar(10);default:'user'" json:"role"`
	CreatedAt time.Time      `gorm:"<-:create" json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"` // json:"-" artinya tidak akan muncul di JSON response
}
