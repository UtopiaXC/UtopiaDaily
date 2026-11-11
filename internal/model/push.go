package model

import "time"

type UserPushSubscription struct {
	ID            uint             `gorm:"primaryKey" json:"id"`
	UserID        uint             `gorm:"not null;uniqueIndex:idx_user_push_mod" json:"user_id"`
	User          *User            `gorm:"foreignKey:UserID;references:ID;constraint:OnDelete:CASCADE;" json:"-"`
	PushModuleID  string           `gorm:"size:128;not null;uniqueIndex:idx_user_push_mod" json:"push_module_id"`
	PushModule    *Module          `gorm:"foreignKey:PushModuleID;references:ID;constraint:OnDelete:CASCADE;" json:"-"`
	Enabled       bool             `gorm:"default:false;not null" json:"enabled"`
	Target        string           `gorm:"size:255" json:"target"`
	CronExpr      string           `gorm:"size:64;default:0 8 * * *;not null" json:"cron_expr"`
	DigestTitle   string           `gorm:"size:128" json:"digest_title"`
	MaxItems      int              `gorm:"default:25;not null" json:"max_items"`
	LookbackHours int              `gorm:"default:24;not null" json:"lookback_hours"`
	LastPushedAt  *time.Time       `json:"last_pushed_at,omitempty"`
	Sources       []UserPushSource `gorm:"foreignKey:SubscriptionID;constraint:OnDelete:CASCADE;" json:"sources,omitempty"`
	CreatedAt     time.Time        `json:"created_at"`
	UpdatedAt     time.Time        `json:"updated_at"`
}

func (UserPushSubscription) TableName() string {
	return "user_push_subscriptions"
}

type UserPushSource struct {
	SubscriptionID  uint   `gorm:"primaryKey" json:"subscription_id"`
	ScraperModuleID string `gorm:"primaryKey;size:128" json:"scraper_module_id"`
}

func (UserPushSource) TableName() string {
	return "user_push_sources"
}

type PushDeliveryLog struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	UserID       *uint     `gorm:"index" json:"user_id,omitempty"`
	PushModuleID string    `gorm:"size:128;not null;index" json:"push_module_id"`
	Recipient    string    `gorm:"size:255;not null" json:"recipient"`
	Status       string    `gorm:"size:32;not null;index" json:"status"`
	ItemCount    int       `gorm:"default:0;not null" json:"item_count"`
	ErrorCode    string    `gorm:"size:128" json:"error_code,omitempty"`
	Details      string    `gorm:"type:text" json:"details,omitempty"`
	CreatedAt    time.Time `gorm:"index;not null" json:"created_at"`
}

func (PushDeliveryLog) TableName() string {
	return "push_delivery_logs"
}
