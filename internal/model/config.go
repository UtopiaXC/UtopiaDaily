package model

import "time"

type SystemConfig struct {
	Key       string    `gorm:"primaryKey;size:128" json:"key"`
	Value     string    `gorm:"type:text" json:"value"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (SystemConfig) TableName() string {
	return "system_configs"
}
