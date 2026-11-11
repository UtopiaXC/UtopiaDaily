package model

import "time"

type NewspaperModuleConfig struct {
	ModuleID             string    `gorm:"primaryKey;size:128" json:"module_id"`
	Module               *Module   `gorm:"foreignKey:ModuleID;references:ID;constraint:OnDelete:CASCADE;" json:"-"`
	Visible              bool      `gorm:"default:true;not null" json:"visible"`
	LookbackDaysOverride int       `gorm:"default:0;not null" json:"lookback_days_override"`
	MaxItemsOverride     int       `gorm:"default:0;not null" json:"max_items_override"`
	SortOrder            int       `gorm:"default:0;not null" json:"sort_order"`
	UpdatedAt            time.Time `json:"updated_at"`
}

func (NewspaperModuleConfig) TableName() string {
	return "newspaper_module_configs"
}
