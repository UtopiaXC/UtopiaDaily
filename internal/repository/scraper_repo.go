package repository

import (
	"time"

	"github.com/UtopiaXC/UtopiaDaily/internal/classification"
	"github.com/UtopiaXC/UtopiaDaily/internal/model"
	"github.com/UtopiaXC/UtopiaDaily/internal/module"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ScraperRepository struct {
	db *gorm.DB
}

func NewScraperRepository(db *gorm.DB) *ScraperRepository {
	return &ScraperRepository{db: db}
}

func (r *ScraperRepository) SaveScrapedItems(moduleID, taskKey string, items []module.ScrapedItem) (int, []model.ScraperResult, error) {
	if len(items) == 0 {
		return 0, nil, nil
	}
	now := time.Now().UTC()
	var saved []model.ScraperResult

	err := r.db.Transaction(func(tx *gorm.DB) error {
		for _, item := range items {
			if item.Fingerprint == "" {
				continue
			}
			var count int64
			if err := tx.Model(&model.ScraperResult{}).Where("fingerprint = ?", item.Fingerprint).Count(&count).Error; err != nil {
				return err
			}
			if count > 0 {
				continue
			}

			ctype := item.ContentType
			if ctype == "" {
				ctype = "markdown"
			}
			renderType := item.RenderType
			if renderType == "" {
				if ctype == "html" {
					renderType = "html"
				} else {
					renderType = "raw"
				}
			}
			pubTime := item.PublishedAt
			if pubTime.IsZero() {
				pubTime = now
			}

			rec := model.ScraperResult{
				ModuleID:         moduleID,
				TaskKey:          taskKey,
				Fingerprint:      item.Fingerprint,
				Title:            item.Title,
				Summary:          item.Summary,
				Source:           item.SourceName,
				Author:           item.Author,
				FromURL:          item.ExternalURL,
				Content:          item.Content,
				ContentType:      ctype,
				RenderType:       renderType,
				DatetimeReleased: pubTime,
				CreatedAt:        now,
			}
			if item.Quotation != nil {
				rec.QuotationContent = item.Quotation.Content
				rec.QuotationFrom = item.Quotation.Author
				rec.QuotationURL = item.Quotation.URL
			}

			if err := tx.Create(&rec).Error; err != nil {
				return err
			}

			for idx, m := range item.Media {
				if m.URL == "" {
					continue
				}
				mType := m.Type
				if mType == "" {
					mType = "image"
				}
				att := model.ScraperResultAttachment{
					ResultID:    rec.ID,
					OriginalURL: m.URL,
					MediaType:   mType,
					StorePolicy: "store_and_replace",
					SortOrder:   idx,
				}
				if err := tx.Create(&att).Error; err != nil {
					return err
				}
			}
			saved = append(saved, rec)
		}
		return nil
	})
	return len(saved), saved, err
}

func (r *ScraperRepository) AttachTags(resultID uint, tags []classification.TagAssignment) error {
	if len(tags) == 0 {
		return nil
	}
	now := time.Now().UTC()
	return r.db.Transaction(func(tx *gorm.DB) error {
		for _, ta := range tags {
			var tag model.Tag
			err := tx.Where("group_name = ? AND name = ?", ta.GroupName, ta.TagName).First(&tag).Error
			if err != nil {
				tag = model.Tag{
					GroupName: ta.GroupName,
					Name:      ta.TagName,
					CreatedAt: now,
				}
				if err := tx.Create(&tag).Error; err != nil {
					return err
				}
			}
			rt := model.ScraperResultTag{
				ResultID:   resultID,
				TagID:      tag.ID,
				Confidence: ta.Confidence,
			}
			if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&rt).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *ScraperRepository) QueryRecent(lookbackDays, limit int, moduleIDs []string) ([]model.ScraperResult, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	q := r.db.Model(&model.ScraperResult{}).
		Preload("Attachments", func(tx *gorm.DB) *gorm.DB {
			return tx.Order("sort_order ASC, id ASC")
		}).
		Preload("ResultTags.Tag")

	if lookbackDays > 0 {
		cutoff := time.Now().UTC().AddDate(0, 0, -lookbackDays)
		q = q.Where("datetime_released >= ?", cutoff)
	}
	if len(moduleIDs) > 0 {
		q = q.Where("module_id IN ?", moduleIDs)
	}

	var list []model.ScraperResult
	err := q.Order("datetime_released DESC, id DESC").Limit(limit).Find(&list).Error
	return list, err
}

func (r *ScraperRepository) QuerySince(since time.Time, limit int, moduleIDs []string) ([]model.ScraperResult, error) {
	if limit <= 0 || limit > 200 {
		limit = 30
	}
	q := r.db.Model(&model.ScraperResult{}).
		Where("created_at >= ?", since)
	if len(moduleIDs) > 0 {
		q = q.Where("module_id IN ?", moduleIDs)
	}
	var list []model.ScraperResult
	err := q.Order("datetime_released DESC, id DESC").Limit(limit).Find(&list).Error
	return list, err
}

func (r *ScraperRepository) ListDistinctSources() ([]string, error) {
	var sources []string
	err := r.db.Model(&model.ScraperResult{}).Distinct("source").Order("source ASC").Pluck("source", &sources).Error
	return sources, err
}

func (r *ScraperRepository) CountResults() (int64, error) {
	var cnt int64
	err := r.db.Model(&model.ScraperResult{}).Count(&cnt).Error
	return cnt, err
}

func (r *ScraperRepository) CountClassificationSlot() (groups int64, tags int64) {
	_ = r.db.Model(&model.Tag{}).Distinct("group_name").Count(&groups).Error
	_ = r.db.Model(&model.Tag{}).Count(&tags).Error
	return groups, tags
}
