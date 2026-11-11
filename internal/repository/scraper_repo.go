package repository

import (
	"strings"
	"time"

	"github.com/UtopiaXC/UtopiaDaily/internal/classification"
	"github.com/UtopiaXC/UtopiaDaily/internal/model"
	"github.com/UtopiaXC/UtopiaDaily/internal/module"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type DataQueryParams struct {
	Page         int
	PageSize     int
	Search       string
	ModuleID     string
	TagID        uint
	UntaggedOnly bool
	DateFrom     string
	DateTo       string
}

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
			tagID := ta.TagID
			if tagID == 0 {
				var found model.Tag
				err := tx.Joins("JOIN tag_groups ON tag_groups.id = tags.group_id").
					Where("tag_groups.name = ? AND tags.name = ?", ta.GroupName, ta.TagName).
					First(&found).Error
				if err != nil {
					continue
				}
				tagID = found.ID
			}
			prov := ta.Provider
			if prov == "" {
				prov = "auto"
			}
			rt := model.ScraperResultTag{
				ResultID:   resultID,
				TagID:      tagID,
				Confidence: ta.Confidence,
				Provider:   prov,
				Reason:     ta.Reason,
				CreatedAt:  now,
			}
			if err := tx.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "result_id"}, {Name: "tag_id"}},
				DoUpdates: clause.AssignmentColumns([]string{"confidence", "provider", "reason", "created_at"}),
			}).Create(&rt).Error; err != nil {
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
		Preload("ResultTags.Tag.Group")

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

func (r *ScraperRepository) QueryByModule(moduleID string, lookbackDays, limit int) ([]model.ScraperResult, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	q := r.db.Model(&model.ScraperResult{}).
		Where("module_id = ?", moduleID).
		Preload("Attachments", func(tx *gorm.DB) *gorm.DB {
			return tx.Order("sort_order ASC, id ASC")
		}).
		Preload("ResultTags.Tag.Group")
	if lookbackDays > 0 {
		cutoff := time.Now().UTC().AddDate(0, 0, -lookbackDays)
		q = q.Where("datetime_released >= ?", cutoff)
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

func (r *ScraperRepository) QueryPaginated(p DataQueryParams) ([]model.ScraperResult, int64, error) {
	if p.Page <= 0 {
		p.Page = 1
	}
	if p.PageSize <= 0 || p.PageSize > 100 {
		p.PageSize = 20
	}

	q := r.db.Model(&model.ScraperResult{})
	if s := strings.TrimSpace(p.Search); s != "" {
		like := "%" + s + "%"
		q = q.Where("title LIKE ? OR summary LIKE ? OR source LIKE ? OR author LIKE ?", like, like, like, like)
	}
	if m := strings.TrimSpace(p.ModuleID); m != "" && m != "all" {
		q = q.Where("module_id = ?", m)
	}
	if p.UntaggedOnly {
		q = q.Where("NOT EXISTS (SELECT 1 FROM scraper_result_tags WHERE scraper_result_tags.result_id = scraper_results.id)")
	} else if p.TagID > 0 {
		q = q.Where("EXISTS (SELECT 1 FROM scraper_result_tags WHERE scraper_result_tags.result_id = scraper_results.id AND scraper_result_tags.tag_id = ?)", p.TagID)
	}
	if df := strings.TrimSpace(p.DateFrom); df != "" {
		if t, err := time.Parse("2006-01-02", df); err == nil {
			q = q.Where("datetime_released >= ?", t.UTC())
		}
	}
	if dt := strings.TrimSpace(p.DateTo); dt != "" {
		if t, err := time.Parse("2006-01-02", dt); err == nil {
			q = q.Where("datetime_released < ?", t.AddDate(0, 0, 1).UTC())
		}
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var list []model.ScraperResult
	err := q.
		Preload("Attachments", func(tx *gorm.DB) *gorm.DB {
			return tx.Order("sort_order ASC, id ASC")
		}).
		Preload("ResultTags.Tag.Group").
		Order("datetime_released DESC, id DESC").
		Offset((p.Page - 1) * p.PageSize).
		Limit(p.PageSize).
		Find(&list).Error
	return list, total, err
}

func (r *ScraperRepository) GetDetailByID(id uint) (*model.ScraperResult, error) {
	var rec model.ScraperResult
	err := r.db.
		Preload("Attachments", func(tx *gorm.DB) *gorm.DB {
			return tx.Order("sort_order ASC, id ASC")
		}).
		Preload("ResultTags.Tag.Group").
		Where("id = ?", id).
		First(&rec).Error
	if err != nil {
		return nil, err
	}
	return &rec, nil
}

func (r *ScraperRepository) GetBatchModuleMap(ids []uint) ([]struct {
	ResultID uint
	ModuleID string
}, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	type row struct {
		ID       uint
		ModuleID string
	}
	var rows []row
	if err := r.db.Model(&model.ScraperResult{}).Select("id, module_id").Where("id IN ?", ids).Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]struct {
		ResultID uint
		ModuleID string
	}, 0, len(rows))
	for _, rw := range rows {
		out = append(out, struct {
			ResultID uint
			ModuleID string
		}{ResultID: rw.ID, ModuleID: rw.ModuleID})
	}
	return out, nil
}

func (r *ScraperRepository) DeleteByIDs(ids []uint) (int64, error) {
	if len(ids) == 0 {
		return 0, nil
	}
	var deleted int64
	err := r.db.Transaction(func(tx *gorm.DB) error {
		_ = tx.Where("result_id IN ?", ids).Delete(&model.ScraperResultTag{}).Error
		_ = tx.Where("result_id IN ?", ids).Delete(&model.ScraperResultAttachment{}).Error
		_ = tx.Where("result_id IN ?", ids).Delete(&model.ClassificationTask{}).Error
		res := tx.Where("id IN ?", ids).Delete(&model.ScraperResult{})
		deleted = res.RowsAffected
		return res.Error
	})
	return deleted, err
}

func (r *ScraperRepository) ListDistinctSources() ([]string, error) {
	var sources []string
	err := r.db.Model(&model.ScraperResult{}).Distinct("source").Order("source ASC").Pluck("source", &sources).Error
	return sources, err
}

func (r *ScraperRepository) ListDistinctModuleIDs() ([]string, error) {
	var mods []string
	err := r.db.Model(&model.ScraperResult{}).Distinct("module_id").Order("module_id ASC").Pluck("module_id", &mods).Error
	return mods, err
}

func (r *ScraperRepository) CountResults() (int64, error) {
	var cnt int64
	err := r.db.Model(&model.ScraperResult{}).Count(&cnt).Error
	return cnt, err
}

func (r *ScraperRepository) CountClassificationSlot() (groups int64, tags int64) {
	var defProfile model.TagProfile
	if err := r.db.Order("is_default DESC, id ASC").First(&defProfile).Error; err == nil && defProfile.ID > 0 {
		_ = r.db.Model(&model.TagGroup{}).Where("profile_id = ?", defProfile.ID).Count(&groups).Error
		_ = r.db.Model(&model.Tag{}).Joins("JOIN tag_groups ON tag_groups.id = tags.group_id").Where("tag_groups.profile_id = ?", defProfile.ID).Count(&tags).Error
		return groups, tags
	}
	_ = r.db.Model(&model.TagGroup{}).Count(&groups).Error
	_ = r.db.Model(&model.Tag{}).Count(&tags).Error
	return groups, tags
}
