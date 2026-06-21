package models

import (
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

type PriceCatalog struct {
	ID         uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	VendorName string    `gorm:"type:varchar(100);not null;uniqueIndex:idx_pc_vendor_item" json:"vendor_name"`
	ItemName   string    `gorm:"type:varchar(200);not null;uniqueIndex:idx_pc_vendor_item" json:"item_name"`
	Price      int       `gorm:"not null" json:"price"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type SummaryData struct {
	Vendors        []SummaryVendor `json:"vendors"`
	TotalEstimated int             `json:"total_estimated"`
}

type SummaryVendor struct {
	Name  string            `json:"name"`
	Items []SummaryItemData `json:"items"`
}

type SummaryItemData struct {
	Name            string `json:"name"`
	Qty             int    `json:"qty"`
	UnitPrice       int    `json:"unit_price"`
	IsEstimated     bool   `json:"is_estimated"`
	UnitPriceActual *int   `json:"unit_price_actual,omitempty"`
}

type ListmakSummary struct {
	ID             uint         `gorm:"primaryKey;autoIncrement" json:"id"`
	ListmakID      uint         `gorm:"not null;uniqueIndex" json:"listmak_id"`
	SummaryRaw     string       `gorm:"column:summary_data;type:jsonb;not null" json:"-"`
	Summary        *SummaryData `gorm:"-" json:"summary"`
	OrderWatermark uint         `gorm:"not null" json:"order_watermark"`
	TotalEstimated int          `json:"total_estimated"`
	TotalActual    *int         `json:"total_actual"`
	GeneratedAt    time.Time    `json:"generated_at"`
	ConfirmedAt    *time.Time   `json:"confirmed_at"`
}

func (s *ListmakSummary) AfterFind(tx *gorm.DB) error {
	if s.SummaryRaw == "" {
		return nil
	}
	var data SummaryData
	if err := json.Unmarshal([]byte(s.SummaryRaw), &data); err != nil {
		return err
	}
	s.Summary = &data
	return nil
}

func (s *ListmakSummary) BeforeSave(tx *gorm.DB) error {
	if s.Summary == nil {
		return nil
	}
	b, err := json.Marshal(s.Summary)
	if err != nil {
		return err
	}
	s.SummaryRaw = string(b)
	s.TotalEstimated = s.Summary.TotalEstimated
	return nil
}
