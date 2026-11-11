package model

import "time"

type ScraperResult struct {
	ID               uint                      `gorm:"primaryKey" json:"id"`
	ModuleID         string                    `gorm:"size:128;not null;index" json:"module_id"`
	Module           *Module                   `gorm:"foreignKey:ModuleID;references:ID;constraint:OnDelete:CASCADE;" json:"-"`
	TaskKey          string                    `gorm:"size:80;not null;default:''" json:"task_key,omitempty"`
	Fingerprint      string                    `gorm:"size:255;uniqueIndex;not null" json:"fingerprint"`
	Title            string                    `gorm:"size:512;not null" json:"title"`
	Summary          string                    `gorm:"type:text" json:"summary"`
	Source           string                    `gorm:"size:255;not null" json:"source"`
	Author           string                    `gorm:"size:128" json:"author,omitempty"`
	FromURL          string                    `gorm:"size:1024" json:"from_url"`
	Content          string                    `gorm:"type:text" json:"content"`
	ContentType      string                    `gorm:"size:32;default:markdown;not null" json:"content_type"`
	RenderType       string                    `gorm:"size:32;default:raw;not null" json:"render_type"`
	QuotationContent string                    `gorm:"type:text" json:"quotation_content,omitempty"`
	QuotationFrom    string                    `gorm:"size:255" json:"quotation_from,omitempty"`
	QuotationURL     string                    `gorm:"size:1024" json:"quotation_url,omitempty"`
	QuotationAt      string                    `gorm:"size:64" json:"quotation_at,omitempty"`
	MetadataJSON     string                    `gorm:"type:text" json:"metadata_json,omitempty"`
	DatetimeReleased time.Time                 `gorm:"index;not null" json:"datetime_released"`
	CreatedAt        time.Time                 `gorm:"index;not null" json:"created_at"`
	Attachments      []ScraperResultAttachment `gorm:"foreignKey:ResultID;constraint:OnDelete:CASCADE;" json:"attachments,omitempty"`
	ResultTags       []ScraperResultTag        `gorm:"foreignKey:ResultID;constraint:OnDelete:CASCADE;" json:"result_tags,omitempty"`
}

func (ScraperResult) TableName() string {
	return "scraper_results"
}

type ScraperResultAttachment struct {
	ID          uint   `gorm:"primaryKey" json:"id"`
	ResultID    uint   `gorm:"not null;index" json:"result_id"`
	OriginalURL string `gorm:"size:1024;not null" json:"original_url"`
	CachedPath  string `gorm:"size:512" json:"cached_path,omitempty"`
	MediaType   string `gorm:"size:32;default:image;not null" json:"type"`
	StorePolicy string `gorm:"size:32;default:store_and_replace;not null" json:"store_policy"`
	SortOrder   int    `gorm:"default:0;not null" json:"sort_order"`
}

func (ScraperResultAttachment) TableName() string {
	return "scraper_result_attachments"
}
