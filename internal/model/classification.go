package model

import "time"

type TagProfile struct {
	ID          uint       `gorm:"primaryKey" json:"id"`
	Name        string     `gorm:"size:128;uniqueIndex;not null" json:"name"`
	Description string     `gorm:"size:512" json:"description,omitempty"`
	IsDefault   bool       `gorm:"default:false;not null" json:"is_default"`
	Groups      []TagGroup `gorm:"foreignKey:ProfileID;constraint:OnDelete:CASCADE;" json:"groups,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

func (TagProfile) TableName() string {
	return "tag_profiles"
}

type TagGroup struct {
	ID            uint        `gorm:"primaryKey" json:"id"`
	ProfileID     uint        `gorm:"not null;uniqueIndex:idx_profile_group" json:"profile_id"`
	Profile       *TagProfile `gorm:"foreignKey:ProfileID;references:ID;constraint:OnDelete:CASCADE;" json:"-"`
	Name          string      `gorm:"size:64;not null;uniqueIndex:idx_profile_group" json:"name"`
	Description   string      `gorm:"size:512" json:"description,omitempty"`
	SelectionMode string      `gorm:"size:16;default:limit;not null" json:"selection_mode"`
	MaxResults    int         `gorm:"default:1;not null" json:"max_results"`
	SortOrder     int         `gorm:"default:0;not null" json:"sort_order"`
	Tags          []Tag       `gorm:"foreignKey:GroupID;constraint:OnDelete:CASCADE;" json:"tags,omitempty"`
}

func (TagGroup) TableName() string {
	return "tag_groups"
}

type Tag struct {
	ID          uint         `gorm:"primaryKey" json:"id"`
	GroupID     uint         `gorm:"not null;uniqueIndex:idx_group_tag" json:"group_id"`
	Group       *TagGroup    `gorm:"foreignKey:GroupID;references:ID;constraint:OnDelete:CASCADE;" json:"group,omitempty"`
	Name        string       `gorm:"size:64;not null;uniqueIndex:idx_group_tag" json:"name"`
	Description string       `gorm:"size:512" json:"description,omitempty"`
	Priority    int          `gorm:"default:0;not null" json:"priority"`
	SortOrder   int          `gorm:"default:0;not null" json:"sort_order"`
	Keywords    []TagKeyword `gorm:"foreignKey:TagID;constraint:OnDelete:CASCADE;" json:"keywords,omitempty"`
	CreatedAt   time.Time    `json:"created_at"`
}

func (Tag) TableName() string {
	return "tags"
}

type TagKeyword struct {
	ID      uint   `gorm:"primaryKey" json:"id"`
	TagID   uint   `gorm:"not null;uniqueIndex:idx_tag_keyword" json:"tag_id"`
	Keyword string `gorm:"size:128;not null;uniqueIndex:idx_tag_keyword" json:"keyword"`
}

func (TagKeyword) TableName() string {
	return "tag_keywords"
}

type ScraperResultTag struct {
	ResultID   uint      `gorm:"primaryKey" json:"result_id"`
	TagID      uint      `gorm:"primaryKey" json:"tag_id"`
	Tag        *Tag      `gorm:"foreignKey:TagID;references:ID;constraint:OnDelete:CASCADE;" json:"tag,omitempty"`
	Confidence float64   `gorm:"default:1;not null" json:"confidence"`
	Provider   string    `gorm:"size:32;default:manual;not null" json:"provider"`
	Reason     string    `gorm:"size:512" json:"reason,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
}

func (ScraperResultTag) TableName() string {
	return "scraper_result_tags"
}

type ModuleClassificationConfig struct {
	ModuleID         string      `gorm:"primaryKey;size:128" json:"module_id"`
	Module           *Module     `gorm:"foreignKey:ModuleID;references:ID;constraint:OnDelete:CASCADE;" json:"-"`
	OverrideSystem   bool        `gorm:"default:false;not null" json:"override_system"`
	Provider         string      `gorm:"size:32;default:none;not null" json:"provider"`
	ModelName        string      `gorm:"size:255" json:"model_name"`
	Confidence       int         `gorm:"default:60;not null" json:"confidence"`
	MaxInputChars    *int        `json:"max_input_chars,omitempty"`
	TruncateStrategy string      `gorm:"size:32" json:"truncate_strategy"`
	ProfileID        *uint       `gorm:"index" json:"profile_id,omitempty"`
	Profile          *TagProfile `gorm:"foreignKey:ProfileID;references:ID;constraint:OnDelete:SET NULL;" json:"-"`
	LLMProtocol      string      `gorm:"size:32;default:ollama;not null" json:"llm_protocol"`
	LLMURL           string      `gorm:"size:512" json:"llm_url"`
	LLMAPIKey        string      `gorm:"size:512" json:"llm_api_key,omitempty"`
	LLMModel         string      `gorm:"size:255" json:"llm_model"`
	LLMPrompt        string      `gorm:"type:text" json:"llm_prompt,omitempty"`
	UpdatedAt        time.Time   `json:"updated_at"`
}

func (ModuleClassificationConfig) TableName() string {
	return "module_classification_configs"
}

const (
	ClassTaskPending   = "pending"
	ClassTaskRunning   = "running"
	ClassTaskCompleted = "completed"
	ClassTaskFailed    = "failed"
	ClassTaskCancelled = "cancelled"
	ClassTaskSkipped   = "skipped"
)

type ClassificationTask struct {
	ID                uint           `gorm:"primaryKey" json:"id"`
	ResultID          uint           `gorm:"not null;index" json:"result_id"`
	Result            *ScraperResult `gorm:"foreignKey:ResultID;references:ID;constraint:OnDelete:CASCADE;" json:"result,omitempty"`
	ModuleID          string         `gorm:"size:128;not null;index" json:"module_id"`
	Status            string         `gorm:"size:24;not null;index" json:"status"`
	Provider          string         `gorm:"size:32" json:"provider"`
	ModelName         string         `gorm:"size:255" json:"model_name"`
	Attempts          int            `gorm:"default:0;not null" json:"attempts"`
	ErrorMessage      string         `gorm:"type:text" json:"error_message,omitempty"`
	AssignedTagsCount int            `gorm:"default:0;not null" json:"assigned_tags_count"`
	CreatedAt         time.Time      `gorm:"index;not null" json:"created_at"`
	StartedAt         *time.Time     `json:"started_at,omitempty"`
	FinishedAt        *time.Time     `json:"finished_at,omitempty"`
}

func (ClassificationTask) TableName() string {
	return "classification_tasks"
}
