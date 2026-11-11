package repository

import (
	"errors"
	"strings"
	"time"

	"github.com/UtopiaXC/UtopiaDaily/internal/classification"
	"github.com/UtopiaXC/UtopiaDaily/internal/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type QueueStats struct {
	Pending   int64 `json:"pending"`
	Running   int64 `json:"running"`
	Completed int64 `json:"completed"`
	Failed    int64 `json:"failed"`
	Cancelled int64 `json:"cancelled"`
	Skipped   int64 `json:"skipped"`
}

type ClassificationTaskView struct {
	ID                uint       `json:"id"`
	ResultID          uint       `json:"result_id"`
	ArticleTitle      string     `json:"article_title"`
	ArticleSource     string     `json:"article_source"`
	ModuleID          string     `json:"module_id"`
	Status            string     `json:"status"`
	Provider          string     `json:"provider"`
	ModelName         string     `json:"model_name"`
	Attempts          int        `json:"attempts"`
	ErrorMessage      string     `json:"error_message,omitempty"`
	AssignedTagsCount int        `json:"assigned_tags_count"`
	CreatedAt         time.Time  `json:"created_at"`
	StartedAt         *time.Time `json:"started_at,omitempty"`
	FinishedAt        *time.Time `json:"finished_at,omitempty"`
}

type AssignedTagDetail struct {
	TagID      uint    `json:"tag_id"`
	GroupName  string  `json:"group_name"`
	TagName    string  `json:"tag_name"`
	Confidence float64 `json:"confidence"`
	Provider   string  `json:"provider"`
	Reason     string  `json:"reason,omitempty"`
}

type ClassificationTaskArticleView struct {
	ID          uint      `json:"id"`
	Title       string    `json:"title"`
	Summary     string    `json:"summary"`
	Content     string    `json:"content"`
	ContentType string    `json:"content_type"`
	RenderType  string    `json:"render_type"`
	Source      string    `json:"source"`
	Author      string    `json:"author,omitempty"`
	FromURL     string    `json:"from_url,omitempty"`
	PublishedAt time.Time `json:"datetime_released"`
}

type ClassificationTaskDetailView struct {
	ClassificationTaskView
	Article      *ClassificationTaskArticleView `json:"article,omitempty"`
	AssignedTags []AssignedTagDetail            `json:"assigned_tags"`
}

type ProfileGroupInput struct {
	ID            uint              `json:"id"`
	Name          string            `json:"name"`
	Description   string            `json:"description"`
	SelectionMode string            `json:"selection_mode"`
	MaxResults    int               `json:"max_results"`
	Tags          []ProfileTagInput `json:"tags"`
}

type ProfileTagInput struct {
	ID          uint     `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Priority    int      `json:"priority"`
	Keywords    []string `json:"keywords"`
}

type ProfileSaveInput struct {
	Name        string              `json:"name"`
	Description string              `json:"description"`
	IsDefault   bool                `json:"is_default"`
	Groups      []ProfileGroupInput `json:"groups"`
}

type ClassificationRepository struct {
	db *gorm.DB
}

func NewClassificationRepository(db *gorm.DB) *ClassificationRepository {
	return &ClassificationRepository{db: db}
}

func (r *ClassificationRepository) ListProfiles() ([]model.TagProfile, error) {
	var profiles []model.TagProfile
	err := r.db.
		Preload("Groups", func(tx *gorm.DB) *gorm.DB {
			return tx.Order("sort_order ASC, id ASC")
		}).
		Preload("Groups.Tags", func(tx *gorm.DB) *gorm.DB {
			return tx.Order("sort_order ASC, id ASC")
		}).
		Preload("Groups.Tags.Keywords", func(tx *gorm.DB) *gorm.DB {
			return tx.Order("id ASC")
		}).
		Order("is_default DESC, id ASC").
		Find(&profiles).Error
	return profiles, err
}

func (r *ClassificationRepository) GetProfile(id uint) (*model.TagProfile, error) {
	var p model.TagProfile
	err := r.db.
		Preload("Groups", func(tx *gorm.DB) *gorm.DB {
			return tx.Order("sort_order ASC, id ASC")
		}).
		Preload("Groups.Tags", func(tx *gorm.DB) *gorm.DB {
			return tx.Order("sort_order ASC, id ASC")
		}).
		Preload("Groups.Tags.Keywords", func(tx *gorm.DB) *gorm.DB {
			return tx.Order("id ASC")
		}).
		Where("id = ?", id).
		First(&p).Error
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *ClassificationRepository) GetDefaultProfile() (*model.TagProfile, error) {
	var p model.TagProfile
	err := r.db.
		Preload("Groups", func(tx *gorm.DB) *gorm.DB {
			return tx.Order("sort_order ASC, id ASC")
		}).
		Preload("Groups.Tags", func(tx *gorm.DB) *gorm.DB {
			return tx.Order("sort_order ASC, id ASC")
		}).
		Preload("Groups.Tags.Keywords", func(tx *gorm.DB) *gorm.DB {
			return tx.Order("id ASC")
		}).
		Order("is_default DESC, id ASC").
		First(&p).Error
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func ProfileToSpec(p *model.TagProfile) classification.ProfileSpec {
	if p == nil {
		return classification.ProfileSpec{}
	}
	spec := classification.ProfileSpec{
		ID:          p.ID,
		Name:        p.Name,
		Description: p.Description,
		Groups:      make([]classification.GroupSpec, 0, len(p.Groups)),
	}
	for _, g := range p.Groups {
		gs := classification.GroupSpec{
			ID:            g.ID,
			Name:          g.Name,
			Description:   g.Description,
			SelectionMode: g.SelectionMode,
			MaxResults:    g.MaxResults,
			Tags:          make([]classification.TagSpec, 0, len(g.Tags)),
		}
		for _, t := range g.Tags {
			kws := make([]string, 0, len(t.Keywords))
			for _, k := range t.Keywords {
				if strings.TrimSpace(k.Keyword) != "" {
					kws = append(kws, k.Keyword)
				}
			}
			gs.Tags = append(gs.Tags, classification.TagSpec{
				ID:          t.ID,
				Name:        t.Name,
				Description: t.Description,
				Priority:    t.Priority,
				Keywords:    kws,
			})
		}
		spec.Groups = append(spec.Groups, gs)
	}
	return spec
}

func (r *ClassificationRepository) CreateProfile(input ProfileSaveInput) (*model.TagProfile, error) {
	now := time.Now().UTC()
	var createdID uint
	err := r.db.Transaction(func(tx *gorm.DB) error {
		name := strings.TrimSpace(input.Name)
		if name == "" {
			return errors.New("profile name required")
		}
		if input.IsDefault {
			if err := tx.Model(&model.TagProfile{}).Where("is_default = ?", true).Update("is_default", false).Error; err != nil {
				return err
			}
		}
		p := model.TagProfile{
			Name:        name,
			Description: strings.TrimSpace(input.Description),
			IsDefault:   input.IsDefault,
			CreatedAt:   now,
			UpdatedAt:   now,
		}
		if err := tx.Create(&p).Error; err != nil {
			return err
		}
		if err := syncProfileHierarchyTx(tx, p.ID, input.Groups, now); err != nil {
			return err
		}
		createdID = p.ID
		return nil
	})
	if err != nil {
		return nil, err
	}
	return r.GetProfile(createdID)
}

func (r *ClassificationRepository) UpdateProfile(id uint, input ProfileSaveInput) (*model.TagProfile, error) {
	now := time.Now().UTC()
	err := r.db.Transaction(func(tx *gorm.DB) error {
		var existing model.TagProfile
		if err := tx.Where("id = ?", id).First(&existing).Error; err != nil {
			return err
		}
		name := strings.TrimSpace(input.Name)
		if name == "" {
			name = existing.Name
		}
		if input.IsDefault && !existing.IsDefault {
			if err := tx.Model(&model.TagProfile{}).Where("is_default = ?", true).Update("is_default", false).Error; err != nil {
				return err
			}
		}
		if err := tx.Model(&existing).Updates(map[string]any{
			"name":        name,
			"description": strings.TrimSpace(input.Description),
			"is_default":  input.IsDefault,
			"updated_at":  now,
		}).Error; err != nil {
			return err
		}
		return syncProfileHierarchyTx(tx, id, input.Groups, now)
	})
	if err != nil {
		return nil, err
	}
	return r.GetProfile(id)
}

func (r *ClassificationRepository) DeleteProfile(id uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var total int64
		if err := tx.Model(&model.TagProfile{}).Count(&total).Error; err != nil {
			return err
		}
		if total <= 1 {
			return errors.New("cannot delete last profile")
		}
		var target model.TagProfile
		if err := tx.Where("id = ?", id).First(&target).Error; err != nil {
			return err
		}
		if err := tx.Delete(&target).Error; err != nil {
			return err
		}
		if target.IsDefault {
			var fallback model.TagProfile
			if err := tx.Order("id ASC").First(&fallback).Error; err == nil {
				_ = tx.Model(&fallback).Update("is_default", true).Error
			}
		}
		return nil
	})
}

func syncProfileHierarchyTx(tx *gorm.DB, profileID uint, groups []ProfileGroupInput, now time.Time) error {
	var existingGroups []model.TagGroup
	if err := tx.Where("profile_id = ?", profileID).Find(&existingGroups).Error; err != nil {
		return err
	}
	existingGroupMap := make(map[uint]model.TagGroup, len(existingGroups))
	for _, g := range existingGroups {
		existingGroupMap[g.ID] = g
	}

	keptGroupIDs := make(map[uint]bool)
	for gIdx, gIn := range groups {
		gName := strings.TrimSpace(gIn.Name)
		if gName == "" {
			continue
		}
		mode := strings.ToLower(strings.TrimSpace(gIn.SelectionMode))
		if mode != "single" && mode != "multi" && mode != "limit" {
			mode = "limit"
		}
		maxRes := gIn.MaxResults
		if mode == "single" {
			maxRes = 1
		} else if maxRes <= 0 {
			maxRes = 1
		}

		var groupID uint
		if _, ok := existingGroupMap[gIn.ID]; ok && gIn.ID > 0 {
			groupID = gIn.ID
			if err := tx.Model(&model.TagGroup{}).Where("id = ?", groupID).Updates(map[string]any{
				"name":           gName,
				"description":    strings.TrimSpace(gIn.Description),
				"selection_mode": mode,
				"max_results":    maxRes,
				"sort_order":     gIdx,
			}).Error; err != nil {
				return err
			}
		} else {
			rec := model.TagGroup{
				ProfileID:     profileID,
				Name:          gName,
				Description:   strings.TrimSpace(gIn.Description),
				SelectionMode: mode,
				MaxResults:    maxRes,
				SortOrder:     gIdx,
			}
			if err := tx.Create(&rec).Error; err != nil {
				return err
			}
			groupID = rec.ID
		}
		keptGroupIDs[groupID] = true

		var existingTags []model.Tag
		if err := tx.Where("group_id = ?", groupID).Find(&existingTags).Error; err != nil {
			return err
		}
		existingTagMap := make(map[uint]model.Tag, len(existingTags))
		for _, t := range existingTags {
			existingTagMap[t.ID] = t
		}

		keptTagIDs := make(map[uint]bool)
		for tIdx, tIn := range gIn.Tags {
			tName := strings.TrimSpace(tIn.Name)
			if tName == "" {
				continue
			}
			var tagID uint
			if _, ok := existingTagMap[tIn.ID]; ok && tIn.ID > 0 {
				tagID = tIn.ID
				if err := tx.Model(&model.Tag{}).Where("id = ?", tagID).Updates(map[string]any{
					"group_id":    groupID,
					"name":        tName,
					"description": strings.TrimSpace(tIn.Description),
					"priority":    tIn.Priority,
					"sort_order":  tIdx,
				}).Error; err != nil {
					return err
				}
			} else {
				rec := model.Tag{
					GroupID:     groupID,
					Name:        tName,
					Description: strings.TrimSpace(tIn.Description),
					Priority:    tIn.Priority,
					SortOrder:   tIdx,
					CreatedAt:   now,
				}
				if err := tx.Create(&rec).Error; err != nil {
					return err
				}
				tagID = rec.ID
			}
			keptTagIDs[tagID] = true

			if err := tx.Where("tag_id = ?", tagID).Delete(&model.TagKeyword{}).Error; err != nil {
				return err
			}
			seenKW := make(map[string]bool)
			for _, rawKW := range tIn.Keywords {
				kw := strings.TrimSpace(rawKW)
				if kw == "" || seenKW[kw] {
					continue
				}
				seenKW[kw] = true
				if err := tx.Create(&model.TagKeyword{TagID: tagID, Keyword: kw}).Error; err != nil {
					return err
				}
			}
		}

		for oldTagID := range existingTagMap {
			if !keptTagIDs[oldTagID] {
				if err := tx.Delete(&model.Tag{}, oldTagID).Error; err != nil {
					return err
				}
			}
		}
	}

	for oldGroupID := range existingGroupMap {
		if !keptGroupIDs[oldGroupID] {
			if err := tx.Delete(&model.TagGroup{}, oldGroupID).Error; err != nil {
				return err
			}
		}
	}
	return nil
}

func (r *ClassificationRepository) GetModuleConfig(moduleID string) (*model.ModuleClassificationConfig, error) {
	var cfg model.ModuleClassificationConfig
	err := r.db.Where("module_id = ?", moduleID).First(&cfg).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &model.ModuleClassificationConfig{
				ModuleID:       moduleID,
				OverrideSystem: false,
				Provider:       "none",
				Confidence:     60,
				LLMProtocol:    "ollama",
			}, nil
		}
		return nil, err
	}
	return &cfg, nil
}

func (r *ClassificationRepository) UpsertModuleConfig(cfg *model.ModuleClassificationConfig) error {
	now := time.Now().UTC()
	cfg.UpdatedAt = now
	var existing model.ModuleClassificationConfig
	err := r.db.Where("module_id = ?", cfg.ModuleID).First(&existing).Error
	values := map[string]any{
		"override_system":   cfg.OverrideSystem,
		"provider":          cfg.Provider,
		"model_name":        cfg.ModelName,
		"confidence":        cfg.Confidence,
		"max_input_chars":   cfg.MaxInputChars,
		"truncate_strategy": cfg.TruncateStrategy,
		"profile_id":        cfg.ProfileID,
		"llm_protocol":      cfg.LLMProtocol,
		"llm_url":           cfg.LLMURL,
		"llm_api_key":       cfg.LLMAPIKey,
		"llm_model":         cfg.LLMModel,
		"llm_prompt":        cfg.LLMPrompt,
		"updated_at":        now,
	}
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		values["module_id"] = cfg.ModuleID
		return r.db.Model(&model.ModuleClassificationConfig{}).Create(values).Error
	}
	return r.db.Model(&model.ModuleClassificationConfig{}).Where("module_id = ?", cfg.ModuleID).Updates(values).Error
}

func (r *ClassificationRepository) ListAllTags() ([]model.Tag, error) {
	var tags []model.Tag
	err := r.db.Preload("Group").Preload("Group.Profile").Order("group_id ASC, sort_order ASC, id ASC").Find(&tags).Error
	return tags, err
}

func (r *ClassificationRepository) ReplaceResultTags(resultID uint, tags []classification.TagAssignment, replaceAutoOnly bool) error {
	now := time.Now().UTC()
	return r.db.Transaction(func(tx *gorm.DB) error {
		q := tx.Where("result_id = ?", resultID)
		if replaceAutoOnly {
			q = q.Where("provider != ?", "manual")
		}
		if err := q.Delete(&model.ScraperResultTag{}).Error; err != nil {
			return err
		}
		for _, ta := range tags {
			tagID := ta.TagID
			if tagID == 0 {
				var found model.Tag
				if err := tx.Joins("JOIN tag_groups ON tag_groups.id = tags.group_id").
					Where("tag_groups.name = ? AND tags.name = ?", ta.GroupName, ta.TagName).
					First(&found).Error; err != nil {
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

func (r *ClassificationRepository) AddManualResultTag(resultID, tagID uint) error {
	return r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "result_id"}, {Name: "tag_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"confidence", "provider", "reason", "created_at"}),
	}).Create(&model.ScraperResultTag{
		ResultID:   resultID,
		TagID:      tagID,
		Confidence: 1.0,
		Provider:   "manual",
		Reason:     "manual_assignment",
		CreatedAt:  time.Now().UTC(),
	}).Error
}

func (r *ClassificationRepository) RemoveResultTag(resultID, tagID uint) error {
	return r.db.Where("result_id = ? AND tag_id = ?", resultID, tagID).Delete(&model.ScraperResultTag{}).Error
}

func (r *ClassificationRepository) EnqueueResults(items []struct {
	ResultID uint
	ModuleID string
}) (int, error) {
	if len(items) == 0 {
		return 0, nil
	}
	now := time.Now().UTC()
	added := 0
	err := r.db.Transaction(func(tx *gorm.DB) error {
		for _, it := range items {
			if it.ResultID == 0 {
				continue
			}
			var activeCount int64
			if err := tx.Model(&model.ClassificationTask{}).
				Where("result_id = ? AND status IN ?", it.ResultID, []string{model.ClassTaskPending, model.ClassTaskRunning}).
				Count(&activeCount).Error; err != nil {
				return err
			}
			if activeCount > 0 {
				continue
			}
			t := model.ClassificationTask{
				ResultID:  it.ResultID,
				ModuleID:  it.ModuleID,
				Status:    model.ClassTaskPending,
				CreatedAt: now,
			}
			if err := tx.Create(&t).Error; err != nil {
				return err
			}
			added++
		}
		return nil
	})
	return added, err
}

func (r *ClassificationRepository) ResetStaleRunning() error {
	return r.db.Model(&model.ClassificationTask{}).
		Where("status = ?", model.ClassTaskRunning).
		Updates(map[string]any{
			"status":     model.ClassTaskPending,
			"started_at": nil,
		}).Error
}

func (r *ClassificationRepository) ClaimNextPending() (*model.ClassificationTask, error) {
	var task model.ClassificationTask
	now := time.Now().UTC()
	err := r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Preload("Result").
			Where("status = ?", model.ClassTaskPending).
			Order("id ASC").
			First(&task).Error; err != nil {
			return err
		}
		return tx.Model(&model.ClassificationTask{}).
			Where("id = ?", task.ID).
			Updates(map[string]any{
				"status":     model.ClassTaskRunning,
				"started_at": now,
				"attempts":   task.Attempts + 1,
			}).Error
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	task.Status = model.ClassTaskRunning
	task.StartedAt = &now
	return &task, nil
}

func (r *ClassificationRepository) CompleteTask(taskID uint, provider, modelName string, tagsCount int) error {
	now := time.Now().UTC()
	return r.db.Model(&model.ClassificationTask{}).
		Where("id = ?", taskID).
		Updates(map[string]any{
			"status":              model.ClassTaskCompleted,
			"provider":            provider,
			"model_name":          modelName,
			"assigned_tags_count": tagsCount,
			"error_message":       "",
			"finished_at":         now,
		}).Error
}

func (r *ClassificationRepository) SkipTask(taskID uint, reason string) error {
	now := time.Now().UTC()
	return r.db.Model(&model.ClassificationTask{}).
		Where("id = ?", taskID).
		Updates(map[string]any{
			"status":        model.ClassTaskSkipped,
			"provider":      "none",
			"error_message": reason,
			"finished_at":   now,
		}).Error
}

func (r *ClassificationRepository) FailTask(taskID uint, provider, modelName, errMsg string) error {
	now := time.Now().UTC()
	return r.db.Model(&model.ClassificationTask{}).
		Where("id = ?", taskID).
		Updates(map[string]any{
			"status":              model.ClassTaskFailed,
			"provider":            provider,
			"model_name":          modelName,
			"assigned_tags_count": 0,
			"error_message":       errMsg,
			"finished_at":         now,
		}).Error
}

func (r *ClassificationRepository) CancelPending() (int64, error) {
	now := time.Now().UTC()
	res := r.db.Model(&model.ClassificationTask{}).
		Where("status = ?", model.ClassTaskPending).
		Updates(map[string]any{
			"status":      model.ClassTaskCancelled,
			"finished_at": now,
		})
	return res.RowsAffected, res.Error
}

func (r *ClassificationRepository) CancelTask(taskID uint) error {
	now := time.Now().UTC()
	return r.db.Model(&model.ClassificationTask{}).
		Where("id = ? AND status = ?", taskID, model.ClassTaskPending).
		Updates(map[string]any{
			"status":      model.ClassTaskCancelled,
			"finished_at": now,
		}).Error
}

func (r *ClassificationRepository) RetryTask(taskID uint) error {
	return r.db.Model(&model.ClassificationTask{}).
		Where("id = ? AND status IN ?", taskID, []string{model.ClassTaskFailed, model.ClassTaskCancelled, model.ClassTaskSkipped}).
		Updates(map[string]any{
			"status":        model.ClassTaskPending,
			"error_message": "",
			"started_at":    nil,
			"finished_at":   nil,
		}).Error
}

func (r *ClassificationRepository) RetryFailed() (int64, error) {
	res := r.db.Model(&model.ClassificationTask{}).
		Where("status = ?", model.ClassTaskFailed).
		Updates(map[string]any{
			"status":        model.ClassTaskPending,
			"error_message": "",
			"started_at":    nil,
			"finished_at":   nil,
		})
	return res.RowsAffected, res.Error
}

func (r *ClassificationRepository) ClearFinished() (int64, error) {
	res := r.db.Where("status IN ?", []string{model.ClassTaskCompleted, model.ClassTaskFailed, model.ClassTaskCancelled, model.ClassTaskSkipped}).
		Delete(&model.ClassificationTask{})
	return res.RowsAffected, res.Error
}

func (r *ClassificationRepository) GetQueueStats() (QueueStats, error) {
	var stats QueueStats
	type countRow struct {
		Status string
		Cnt    int64
	}
	var rows []countRow
	if err := r.db.Model(&model.ClassificationTask{}).Select("status, COUNT(*) as cnt").Group("status").Scan(&rows).Error; err != nil {
		return stats, err
	}
	for _, row := range rows {
		switch row.Status {
		case model.ClassTaskPending:
			stats.Pending = row.Cnt
		case model.ClassTaskRunning:
			stats.Running = row.Cnt
		case model.ClassTaskCompleted:
			stats.Completed = row.Cnt
		case model.ClassTaskFailed:
			stats.Failed = row.Cnt
		case model.ClassTaskCancelled:
			stats.Cancelled = row.Cnt
		case model.ClassTaskSkipped:
			stats.Skipped = row.Cnt
		}
	}
	return stats, nil
}

func (r *ClassificationRepository) ListTasks(status string, page, pageSize int) ([]ClassificationTaskView, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}
	q := r.db.Model(&model.ClassificationTask{})
	if status != "" && status != "all" {
		q = q.Where("status = ?", status)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []model.ClassificationTask
	if err := q.Preload("Result").Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	out := make([]ClassificationTaskView, 0, len(rows))
	for _, t := range rows {
		view := ClassificationTaskView{
			ID:                t.ID,
			ResultID:          t.ResultID,
			ModuleID:          t.ModuleID,
			Status:            t.Status,
			Provider:          t.Provider,
			ModelName:         t.ModelName,
			Attempts:          t.Attempts,
			ErrorMessage:      t.ErrorMessage,
			AssignedTagsCount: t.AssignedTagsCount,
			CreatedAt:         t.CreatedAt,
			StartedAt:         t.StartedAt,
			FinishedAt:        t.FinishedAt,
		}
		if t.Result != nil {
			view.ArticleTitle = t.Result.Title
			view.ArticleSource = t.Result.Source
		}
		out = append(out, view)
	}
	return out, total, nil
}

func (r *ClassificationRepository) GetTaskDetail(taskID uint) (*ClassificationTaskDetailView, error) {
	var t model.ClassificationTask
	if err := r.db.Preload("Result").Where("id = ?", taskID).First(&t).Error; err != nil {
		return nil, err
	}

	view := &ClassificationTaskDetailView{
		ClassificationTaskView: ClassificationTaskView{
			ID:                t.ID,
			ResultID:          t.ResultID,
			ModuleID:          t.ModuleID,
			Status:            t.Status,
			Provider:          t.Provider,
			ModelName:         t.ModelName,
			Attempts:          t.Attempts,
			ErrorMessage:      t.ErrorMessage,
			AssignedTagsCount: t.AssignedTagsCount,
			CreatedAt:         t.CreatedAt,
			StartedAt:         t.StartedAt,
			FinishedAt:        t.FinishedAt,
		},
		AssignedTags: make([]AssignedTagDetail, 0),
	}
	if t.Result != nil {
		view.ArticleTitle = t.Result.Title
		view.ArticleSource = t.Result.Source
		view.Article = &ClassificationTaskArticleView{
			ID:          t.Result.ID,
			Title:       t.Result.Title,
			Summary:     t.Result.Summary,
			Content:     t.Result.Content,
			ContentType: t.Result.ContentType,
			RenderType:  t.Result.RenderType,
			Source:      t.Result.Source,
			Author:      t.Result.Author,
			FromURL:     t.Result.FromURL,
			PublishedAt: t.Result.DatetimeReleased,
		}
	}

	var resultTags []model.ScraperResultTag
	if err := r.db.
		Preload("Tag").
		Preload("Tag.Group").
		Where("result_id = ?", t.ResultID).
		Order("confidence DESC, tag_id ASC").
		Find(&resultTags).Error; err != nil {
		return nil, err
	}

	for _, rt := range resultTags {
		detail := AssignedTagDetail{
			TagID:      rt.TagID,
			Confidence: rt.Confidence,
			Provider:   rt.Provider,
			Reason:     rt.Reason,
		}
		if rt.Tag != nil {
			detail.TagName = rt.Tag.Name
			if rt.Tag.Group != nil {
				detail.GroupName = rt.Tag.Group.Name
			}
		}
		view.AssignedTags = append(view.AssignedTags, detail)
	}
	return view, nil
}

func (r *ClassificationRepository) ListUntaggedResults(limit int) ([]struct {
	ResultID uint
	ModuleID string
}, error) {
	if limit <= 0 || limit > 2000 {
		limit = 500
	}
	type row struct {
		ID       uint
		ModuleID string
	}
	var rows []row
	err := r.db.Model(&model.ScraperResult{}).
		Select("scraper_results.id, scraper_results.module_id").
		Where("NOT EXISTS (SELECT 1 FROM scraper_result_tags WHERE scraper_result_tags.result_id = scraper_results.id)").
		Order("scraper_results.datetime_released DESC, scraper_results.id DESC").
		Limit(limit).
		Scan(&rows).Error
	if err != nil {
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
