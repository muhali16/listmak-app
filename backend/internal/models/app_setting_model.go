package models

import "time"

// AppSetting is a simple key-value store for runtime app configuration (e.g.
// testing_mode), togglable by admin without a redeploy.
type AppSetting struct {
	Key       string    `gorm:"primaryKey;type:varchar(50)" json:"key"`
	Value     string    `gorm:"type:text" json:"value"`
	UpdatedAt time.Time `json:"updated_at"`
}
