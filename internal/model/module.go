package model

import "time"

const (
	ModuleKindScraper = "scraper"
	ModuleKindPush    = "push"

	ModuleSourceBuiltin  = "default"
	ModuleSourceExternal = "external"
)

type Module struct {
	ID             string    `gorm:"primaryKey;size:128" json:"module_id"`
	Kind           string    `gorm:"size:32;not null;index" json:"kind"`
	Source         string    `gorm:"size:32;not null;index" json:"source"`
	NameKey        string    `gorm:"size:128;not null" json:"name"`
	DescriptionKey string    `gorm:"size:255" json:"description"`
	Version        string    `gorm:"size:32;not null" json:"version"`
	Author         string    `gorm:"size:64" json:"author"`
	Enabled        bool      `gorm:"default:false;not null" json:"is_enable"`
	EnvReady       bool      `gorm:"default:true;not null" json:"env_ready"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

func (Module) TableName() string {
	return "modules"
}

type ModuleConfig struct {
	ID          uint               `gorm:"primaryKey" json:"id"`
	ModuleID    string             `gorm:"size:128;not null;uniqueIndex:idx_module_config_key" json:"module_id"`
	Module      *Module            `gorm:"foreignKey:ModuleID;references:ID;constraint:OnDelete:CASCADE;" json:"-"`
	ConfigKey   string             `gorm:"size:128;not null;uniqueIndex:idx_module_config_key" json:"config_key"`
	ScalarValue string             `gorm:"type:text" json:"scalar_value"`
	IsArray     bool               `gorm:"default:false;not null" json:"is_array"`
	Items       []ModuleConfigItem `gorm:"foreignKey:ModuleConfigID;constraint:OnDelete:CASCADE;" json:"items,omitempty"`
	UpdatedAt   time.Time          `json:"updated_at"`
}

func (ModuleConfig) TableName() string {
	return "module_configs"
}

type ModuleConfigItem struct {
	ID             uint   `gorm:"primaryKey" json:"id"`
	ModuleConfigID uint   `gorm:"not null;index:idx_config_item_order" json:"module_config_id"`
	ItemOrder      int    `gorm:"not null;index:idx_config_item_order" json:"item_order"`
	ItemValue      string `gorm:"type:text;not null" json:"item_value"`
}

func (ModuleConfigItem) TableName() string {
	return "module_config_items"
}

type ScheduleJob struct {
	ID             uint       `gorm:"primaryKey" json:"id"`
	ModuleID       string     `gorm:"size:128;not null;uniqueIndex:idx_module_task_key" json:"module_id"`
	Module         *Module    `gorm:"foreignKey:ModuleID;references:ID;constraint:OnDelete:CASCADE;" json:"module,omitempty"`
	TaskKey        string     `gorm:"size:128;not null;uniqueIndex:idx_module_task_key" json:"task_key"`
	NameKey        string     `gorm:"size:128;not null" json:"name_key"`
	DescriptionKey string     `gorm:"size:255" json:"description_key"`
	CronExpr       string     `gorm:"size:64;not null" json:"cron_expr"`
	Enabled        bool       `gorm:"default:false;not null" json:"enabled"`
	LastRunAt      *time.Time `json:"last_run_at,omitempty"`
	LastStatus     string     `gorm:"size:32;default:idle;not null" json:"last_status"`
	LastError      string     `gorm:"type:text" json:"last_error,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

func (ScheduleJob) TableName() string {
	return "schedule_jobs"
}
