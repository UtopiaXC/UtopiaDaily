package repository

import (
	"time"

	"github.com/UtopiaXC/UtopiaDaily/internal/model"
	"gorm.io/gorm"
)

type EventQueryParams struct {
	Page     int
	PageSize int
	Level    string
	Category string
	Search   string
	Date     string
	SortBy   string
	Order    string
}

type EventRepository struct {
	db *gorm.DB
}

func NewEventRepository(db *gorm.DB) *EventRepository {
	return &EventRepository{db: db}
}

func (r *EventRepository) Create(ev *model.SystemEvent) error {
	if ev.CreatedAt.IsZero() {
		ev.CreatedAt = time.Now().UTC()
	}
	return r.db.Create(ev).Error
}

func (r *EventRepository) Query(params EventQueryParams) ([]model.SystemEvent, int64, error) {
	if params.Page < 1 {
		params.Page = 1
	}
	if params.PageSize <= 0 || params.PageSize > 100 {
		params.PageSize = 15
	}

	q := r.db.Model(&model.SystemEvent{})
	if params.Level != "" {
		q = q.Where("level = ?", params.Level)
	}
	if params.Category != "" {
		q = q.Where("category = ?", params.Category)
	}
	if params.Date != "" {
		if parsed, err := time.Parse("2006-01-02", params.Date); err == nil {
			start := parsed.UTC()
			end := start.Add(24 * time.Hour)
			q = q.Where("created_at >= ? AND created_at < ?", start, end)
		}
	}
	if params.Search != "" {
		like := "%" + params.Search + "%"
		q = q.Where("event_key LIKE ? OR params_json LIKE ? OR source_id LIKE ?", like, like, like)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	col := "created_at"
	switch params.SortBy {
	case "level", "category", "id", "created_at":
		col = params.SortBy
	}
	dir := "DESC"
	if params.Order == "asc" {
		dir = "ASC"
	}
	orderClause := col + " " + dir + ", id " + dir

	var list []model.SystemEvent
	err := q.Order(orderClause).
		Offset((params.Page - 1) * params.PageSize).
		Limit(params.PageSize).
		Find(&list).Error
	return list, total, err
}

func (r *EventRepository) DistinctCategories() ([]string, error) {
	var cats []string
	err := r.db.Model(&model.SystemEvent{}).Distinct("category").Order("category ASC").Pluck("category", &cats).Error
	return cats, err
}

func (r *EventRepository) Prune(maxCount, maxDays int) error {
	if maxDays > 0 {
		cutoff := time.Now().UTC().AddDate(0, 0, -maxDays)
		if err := r.db.Where("created_at < ?", cutoff).Delete(&model.SystemEvent{}).Error; err != nil {
			return err
		}
	}
	if maxCount > 0 {
		var total int64
		if err := r.db.Model(&model.SystemEvent{}).Count(&total).Error; err == nil && total > int64(maxCount) {
			var cutoffID uint
			if err := r.db.Model(&model.SystemEvent{}).
				Select("id").
				Order("id DESC").
				Offset(maxCount).
				Limit(1).
				Scan(&cutoffID).Error; err == nil && cutoffID > 0 {
				return r.db.Where("id <= ?", cutoffID).Delete(&model.SystemEvent{}).Error
			}
		}
	}
	return nil
}
