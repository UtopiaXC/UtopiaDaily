package repository

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/UtopiaXC/UtopiaDaily/internal/model"
	"github.com/UtopiaXC/UtopiaDaily/internal/module"
	"gorm.io/gorm"
)

type ModuleRepository struct {
	db *gorm.DB
}

func NewModuleRepository(db *gorm.DB) *ModuleRepository {
	return &ModuleRepository{db: db}
}

func (r *ModuleRepository) SyncModule(mod module.Module, envReady bool, customNameKey, customDescKey string) error {
	meta := mod.Meta()
	schema := mod.ConfigSchema()
	tasks := mod.Tasks()
	now := time.Now().UTC()

	nameKey := "module." + meta.ID + ".name"
	if customNameKey != "" {
		nameKey = customNameKey
	}
	descKey := "module." + meta.ID + ".description"
	if customDescKey != "" {
		descKey = customDescKey
	}

	return r.db.Transaction(func(tx *gorm.DB) error {
		var existing model.Module
		err := tx.Where("id = ?", meta.ID).First(&existing).Error
		if err != nil {
			rec := model.Module{
				ID:             meta.ID,
				Kind:           meta.Kind,
				Source:         meta.Source,
				NameKey:        nameKey,
				DescriptionKey: descKey,
				Version:        meta.Version,
				Author:         meta.Author,
				Enabled:        false,
				EnvReady:       envReady,
				CreatedAt:      now,
				UpdatedAt:      now,
			}
			if err := tx.Create(&rec).Error; err != nil {
				return err
			}
		} else {
			if err := tx.Model(&model.Module{}).Where("id = ?", meta.ID).Updates(map[string]any{
				"kind":            meta.Kind,
				"source":          meta.Source,
				"name_key":        nameKey,
				"description_key": descKey,
				"version":         meta.Version,
				"author":          meta.Author,
				"env_ready":       envReady,
				"updated_at":      now,
			}).Error; err != nil {
				return err
			}
		}

		for _, field := range schema {
			var cfgRec model.ModuleConfig
			err := tx.Where("module_id = ? AND config_key = ?", meta.ID, field.Key).First(&cfgRec).Error
			isArray := field.Type == "array"
			if err != nil {
				cfgRec = model.ModuleConfig{
					ModuleID:    meta.ID,
					ConfigKey:   field.Key,
					ScalarValue: field.DefaultValue,
					IsArray:     isArray,
					UpdatedAt:   now,
				}
				if err := tx.Create(&cfgRec).Error; err != nil {
					return err
				}
				if isArray {
					for idx, itemVal := range field.DefaultItems {
						if err := tx.Create(&model.ModuleConfigItem{
							ModuleConfigID: cfgRec.ID,
							ItemOrder:      idx,
							ItemValue:      itemVal,
						}).Error; err != nil {
							return err
						}
					}
				}
			} else if cfgRec.IsArray != isArray {
				if err := tx.Model(&model.ModuleConfig{}).Where("id = ?", cfgRec.ID).Updates(map[string]any{
					"is_array":   isArray,
					"updated_at": now,
				}).Error; err != nil {
					return err
				}
			}
		}

		for _, t := range tasks {
			taskNameKey := fmt.Sprintf("module.%s.task.%s.name", meta.ID, t.Key)
			taskDescKey := fmt.Sprintf("module.%s.task.%s.desc", meta.ID, t.Key)
			var job model.ScheduleJob
			if err := tx.Where("module_id = ? AND task_key = ?", meta.ID, t.Key).First(&job).Error; err != nil {
				cronExpr := t.DefaultCron
				if cronExpr == "" {
					cronExpr = "0 */2 * * *"
				}
				job = model.ScheduleJob{
					ModuleID:       meta.ID,
					TaskKey:        t.Key,
					NameKey:        taskNameKey,
					DescriptionKey: taskDescKey,
					CronExpr:       cronExpr,
					Enabled:        t.DefaultEnabled,
					LastStatus:     "idle",
					CreatedAt:      now,
					UpdatedAt:      now,
				}
				if err := tx.Create(&job).Error; err != nil {
					return err
				}
			} else {
				_ = tx.Model(&model.ScheduleJob{}).Where("id = ?", job.ID).Updates(map[string]any{
					"name_key":        taskNameKey,
					"description_key": taskDescKey,
				}).Error
			}
		}

		return nil
	})
}

func (r *ModuleRepository) ListByKind(kind string) ([]model.Module, error) {
	var list []model.Module
	q := r.db.Model(&model.Module{})
	if kind != "" {
		q = q.Where("kind = ?", kind)
	}
	err := q.Order("source ASC, id ASC").Find(&list).Error
	return list, err
}

func (r *ModuleRepository) GetByID(id string) (*model.Module, error) {
	var m model.Module
	if err := r.db.Where("id = ?", id).First(&m).Error; err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *ModuleRepository) SetEnabled(id string, enabled bool) error {
	return r.db.Model(&model.Module{}).Where("id = ?", id).Updates(map[string]any{
		"enabled":    enabled,
		"updated_at": time.Now().UTC(),
	}).Error
}

func (r *ModuleRepository) SetEnvReady(id string, ready bool) error {
	return r.db.Model(&model.Module{}).Where("id = ?", id).Updates(map[string]any{
		"env_ready":  ready,
		"updated_at": time.Now().UTC(),
	}).Error
}

func (r *ModuleRepository) CountEnabled() (int64, error) {
	var cnt int64
	err := r.db.Model(&model.Module{}).Where("enabled = ?", true).Count(&cnt).Error
	return cnt, err
}

func (r *ModuleRepository) CountEnabledByKind(kind string) (int64, error) {
	var cnt int64
	err := r.db.Model(&model.Module{}).Where("enabled = ? AND kind = ?", true, kind).Count(&cnt).Error
	return cnt, err
}

func (r *ModuleRepository) GetConfigRows(moduleID string) ([]model.ModuleConfig, error) {
	var rows []model.ModuleConfig
	err := r.db.Preload("Items", func(tx *gorm.DB) *gorm.DB {
		return tx.Order("item_order ASC, id ASC")
	}).Where("module_id = ?", moduleID).Order("id ASC").Find(&rows).Error
	return rows, err
}

func (r *ModuleRepository) SaveConfigValues(moduleID string, schema []module.ConfigFieldSpec, payload map[string]any) error {
	specMap := make(map[string]module.ConfigFieldSpec, len(schema))
	for _, s := range schema {
		specMap[s.Key] = s
	}
	now := time.Now().UTC()

	return r.db.Transaction(func(tx *gorm.DB) error {
		for key, rawVal := range payload {
			spec, ok := specMap[key]
			if !ok {
				continue
			}

			var cfgRec model.ModuleConfig
			err := tx.Where("module_id = ? AND config_key = ?", moduleID, key).First(&cfgRec).Error
			if err != nil {
				cfgRec = model.ModuleConfig{
					ModuleID:  moduleID,
					ConfigKey: key,
					IsArray:   spec.Type == "array",
					UpdatedAt: now,
				}
				if err := tx.Create(&cfgRec).Error; err != nil {
					return err
				}
			}

			if spec.Type == "array" {
				if err := tx.Where("module_config_id = ?", cfgRec.ID).Delete(&model.ModuleConfigItem{}).Error; err != nil {
					return err
				}
				items := coerceStringSlice(rawVal)
				order := 0
				for _, itemVal := range items {
					trimmed := strings.TrimSpace(itemVal)
					if trimmed == "" {
						continue
					}
					if err := tx.Create(&model.ModuleConfigItem{
						ModuleConfigID: cfgRec.ID,
						ItemOrder:      order,
						ItemValue:      trimmed,
					}).Error; err != nil {
						return err
					}
					order++
				}
				if err := tx.Model(&model.ModuleConfig{}).Where("id = ?", cfgRec.ID).Updates(map[string]any{
					"is_array":     true,
					"scalar_value": "",
					"updated_at":   now,
				}).Error; err != nil {
					return err
				}
				continue
			}

			strVal := coerceScalarString(rawVal)
			if spec.Secret && strVal == "********" {
				continue
			}
			if err := tx.Model(&model.ModuleConfig{}).Where("id = ?", cfgRec.ID).Updates(map[string]any{
				"is_array":     false,
				"scalar_value": strVal,
				"updated_at":   now,
			}).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *ModuleRepository) ListScheduleJobs() ([]model.ScheduleJob, error) {
	var jobs []model.ScheduleJob
	err := r.db.Preload("Module").Order("module_id ASC, id ASC").Find(&jobs).Error
	return jobs, err
}

func (r *ModuleRepository) GetScheduleJob(id uint) (*model.ScheduleJob, error) {
	var job model.ScheduleJob
	if err := r.db.Preload("Module").Where("id = ?", id).First(&job).Error; err != nil {
		return nil, err
	}
	return &job, nil
}

func (r *ModuleRepository) GetScheduleJobByKey(moduleID, taskKey string) (*model.ScheduleJob, error) {
	var job model.ScheduleJob
	if err := r.db.Preload("Module").Where("module_id = ? AND task_key = ?", moduleID, taskKey).First(&job).Error; err != nil {
		return nil, err
	}
	return &job, nil
}

func (r *ModuleRepository) ListJobsByModule(moduleID string) ([]model.ScheduleJob, error) {
	var jobs []model.ScheduleJob
	err := r.db.Where("module_id = ?", moduleID).Order("id ASC").Find(&jobs).Error
	return jobs, err
}

func (r *ModuleRepository) UpdateScheduleJob(id uint, cronExpr string, enabled bool) error {
	return r.db.Model(&model.ScheduleJob{}).Where("id = ?", id).Updates(map[string]any{
		"cron_expr":  cronExpr,
		"enabled":    enabled,
		"updated_at": time.Now().UTC(),
	}).Error
}

func (r *ModuleRepository) RecordScheduleRun(moduleID, taskKey, status, errMsg string) error {
	now := time.Now().UTC()
	return r.db.Model(&model.ScheduleJob{}).
		Where("module_id = ? AND task_key = ?", moduleID, taskKey).
		Updates(map[string]any{
			"last_run_at": now,
			"last_status": status,
			"last_error":  errMsg,
			"updated_at":  now,
		}).Error
}

type DBConfigAccessor struct {
	specs     map[string]module.ConfigFieldSpec
	rows      map[string]model.ModuleConfig
	sysConfig map[string]string
}

func NewDBConfigAccessor(schema []module.ConfigFieldSpec, rows []model.ModuleConfig, sysConfig map[string]string) *DBConfigAccessor {
	specMap := make(map[string]module.ConfigFieldSpec, len(schema))
	for _, s := range schema {
		specMap[s.Key] = s
	}
	rowMap := make(map[string]model.ModuleConfig, len(rows))
	for _, row := range rows {
		rowMap[row.ConfigKey] = row
	}
	return &DBConfigAccessor{
		specs:     specMap,
		rows:      rowMap,
		sysConfig: sysConfig,
	}
}

func NewMemoryConfigAccessor(schema []module.ConfigFieldSpec, overrides map[string]any, fallbackRows []model.ModuleConfig, sysConfig map[string]string) *DBConfigAccessor {
	base := NewDBConfigAccessor(schema, fallbackRows, sysConfig)
	for _, s := range schema {
		rawVal, ok := overrides[s.Key]
		if !ok {
			continue
		}
		if s.Secret && coerceScalarString(rawVal) == "********" {
			continue
		}
		row := base.rows[s.Key]
		row.ConfigKey = s.Key
		if s.Type == "array" {
			row.IsArray = true
			items := coerceStringSlice(rawVal)
			row.Items = make([]model.ModuleConfigItem, 0, len(items))
			for idx, v := range items {
				if trimmed := strings.TrimSpace(v); trimmed != "" {
					row.Items = append(row.Items, model.ModuleConfigItem{ItemOrder: idx, ItemValue: trimmed})
				}
			}
		} else {
			row.IsArray = false
			row.ScalarValue = coerceScalarString(rawVal)
		}
		base.rows[s.Key] = row
	}
	return base
}

func (a *DBConfigAccessor) GetString(key string) string {
	if r, ok := a.rows[key]; ok {
		return r.ScalarValue
	}
	if s, ok := a.specs[key]; ok {
		return s.DefaultValue
	}
	return ""
}

func (a *DBConfigAccessor) GetInt(key string, fallback int) int {
	val := a.GetString(key)
	if val != "" {
		if n, err := strconv.Atoi(strings.TrimSpace(val)); err == nil {
			return n
		}
	}
	return fallback
}

func (a *DBConfigAccessor) GetBool(key string) bool {
	v := strings.ToLower(strings.TrimSpace(a.GetString(key)))
	return v == "true" || v == "1"
}

func (a *DBConfigAccessor) GetStringSlice(key string) []string {
	if r, ok := a.rows[key]; ok {
		if len(r.Items) > 0 {
			out := make([]string, 0, len(r.Items))
			for _, it := range r.Items {
				out = append(out, it.ItemValue)
			}
			return out
		}
		if r.IsArray {
			return []string{}
		}
		if strings.TrimSpace(r.ScalarValue) != "" {
			lines := strings.Split(r.ScalarValue, "\n")
			out := make([]string, 0, len(lines))
			for _, l := range lines {
				if t := strings.TrimSpace(l); t != "" {
					out = append(out, t)
				}
			}
			return out
		}
	}
	if s, ok := a.specs[key]; ok && len(s.DefaultItems) > 0 {
		out := make([]string, len(s.DefaultItems))
		copy(out, s.DefaultItems)
		return out
	}
	return nil
}

func (a *DBConfigAccessor) RawMap() map[string]any {
	out := make(map[string]any, len(a.specs))
	for key, spec := range a.specs {
		switch spec.Type {
		case "array":
			out[key] = a.GetStringSlice(key)
		case "int":
			out[key] = a.GetInt(key, 0)
		case "bool", "boolean":
			out[key] = a.GetBool(key)
		default:
			out[key] = a.GetString(key)
		}
	}
	return out
}

func (a *DBConfigAccessor) GetSystemConfig(key string) string {
	return a.sysConfig[key]
}

func coerceScalarString(v any) string {
	if v == nil {
		return ""
	}
	switch val := v.(type) {
	case string:
		return val
	case bool:
		if val {
			return "true"
		}
		return "false"
	case float64:
		if val == float64(int64(val)) {
			return strconv.FormatInt(int64(val), 10)
		}
		return strconv.FormatFloat(val, 'f', -1, 64)
	case int:
		return strconv.Itoa(val)
	case int64:
		return strconv.FormatInt(val, 10)
	default:
		return fmt.Sprintf("%v", val)
	}
}

func coerceStringSlice(v any) []string {
	if v == nil {
		return nil
	}
	switch arr := v.(type) {
	case []string:
		return arr
	case []any:
		out := make([]string, 0, len(arr))
		for _, item := range arr {
			out = append(out, coerceScalarString(item))
		}
		return out
	case string:
		lines := strings.Split(arr, "\n")
		out := make([]string, 0, len(lines))
		for _, l := range lines {
			if t := strings.TrimSpace(l); t != "" {
				out = append(out, t)
			}
		}
		return out
	default:
		return nil
	}
}
