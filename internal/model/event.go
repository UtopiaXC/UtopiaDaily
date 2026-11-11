package model

import "time"

const (
	EventLevelNormal   = "NORMAL"
	EventLevelWarning  = "WARNING"
	EventLevelCritical = "CRITICAL"
	EventLevelFatal    = "FATAL"

	EventCategorySystem   = "SYSTEM"
	EventCategoryUser     = "USER"
	EventCategoryModule   = "MODULE"
	EventCategoryTask     = "TASK"
	EventCategorySecurity = "SECURITY"
	EventCategoryPush     = "PUSH"
)

type SystemEvent struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Level       string    `gorm:"size:20;not null;index" json:"level"`
	Category    string    `gorm:"size:32;not null;index" json:"category"`
	EventKey    string    `gorm:"size:128;not null;index" json:"event_key"`
	ParamsJSON  string    `gorm:"type:text" json:"params_json,omitempty"`
	DetailsJSON string    `gorm:"type:text" json:"details_json,omitempty"`
	SourceType  string    `gorm:"size:64" json:"source_type,omitempty"`
	SourceID    string    `gorm:"size:128" json:"source_id,omitempty"`
	CreatedAt   time.Time `gorm:"index;not null" json:"created_at"`
}

func (SystemEvent) TableName() string {
	return "system_events"
}
