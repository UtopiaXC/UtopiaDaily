package model

import "time"

type Tag struct {
	ID        uint         `gorm:"primaryKey" json:"id"`
	GroupName string       `gorm:"size:64;not null;uniqueIndex:idx_tag_group_name" json:"group_name"`
	Name      string       `gorm:"size:64;not null;uniqueIndex:idx_tag_group_name" json:"name"`
	Priority  int          `gorm:"default:0;not null" json:"priority"`
	Keywords  []TagKeyword `gorm:"foreignKey:TagID;constraint:OnDelete:CASCADE;" json:"keywords,omitempty"`
	CreatedAt time.Time    `json:"created_at"`
}

func (Tag) TableName() string {
	return "tags"
}

type TagKeyword struct {
	ID      uint   `gorm:"primaryKey" json:"id"`
	TagID   uint   `gorm:"not null;index" json:"tag_id"`
	Keyword string `gorm:"size:128;not null" json:"keyword"`
}

func (TagKeyword) TableName() string {
	return "tag_keywords"
}

type ScraperResultTag struct {
	ResultID   uint    `gorm:"primaryKey" json:"result_id"`
	TagID      uint    `gorm:"primaryKey" json:"tag_id"`
	Tag        *Tag    `gorm:"foreignKey:TagID;references:ID;constraint:OnDelete:CASCADE;" json:"tag,omitempty"`
	Confidence float64 `gorm:"default:1;not null" json:"confidence"`
}

func (ScraperResultTag) TableName() string {
	return "scraper_result_tags"
}
