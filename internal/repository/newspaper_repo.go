package repository

import (
	"errors"
	"time"

	"github.com/UtopiaXC/UtopiaDaily/internal/model"
	"gorm.io/gorm"
)

type NewspaperModuleConfigDTO struct {
	ModuleID             string `json:"module_id"`
	ModuleName           string `json:"module_name"`
	ModuleEnabled        bool   `json:"module_enabled"`
	Visible              bool   `json:"visible"`
	LookbackDaysOverride int    `json:"lookback_days_override"`
	MaxItemsOverride     int    `json:"max_items_override"`
	SortOrder            int    `json:"sort_order"`
}

type NewspaperModuleConfigInput struct {
	ModuleID             string `json:"module_id"`
	Visible              bool   `json:"visible"`
	LookbackDaysOverride int    `json:"lookback_days_override"`
	MaxItemsOverride     int    `json:"max_items_override"`
	SortOrder            int    `json:"sort_order"`
}

type NewspaperRepository struct {
	db *gorm.DB
}

func NewNewspaperRepository(db *gorm.DB) *NewspaperRepository {
	return &NewspaperRepository{db: db}
}

func (r *NewspaperRepository) ListModuleConfigs() ([]NewspaperModuleConfigDTO, error) {
	var modules []model.Module
	if err := r.db.Where("kind = ?", model.ModuleKindScraper).Order("source ASC, id ASC").Find(&modules).Error; err != nil {
		return nil, err
	}
	var cfgRows []model.NewspaperModuleConfig
	if err := r.db.Find(&cfgRows).Error; err != nil {
		return nil, err
	}
	cfgMap := make(map[string]model.NewspaperModuleConfig, len(cfgRows))
	for _, c := range cfgRows {
		cfgMap[c.ModuleID] = c
	}

	out := make([]NewspaperModuleConfigDTO, 0, len(modules))
	for _, m := range modules {
		dto := NewspaperModuleConfigDTO{
			ModuleID:      m.ID,
			ModuleName:    m.NameKey,
			ModuleEnabled: m.Enabled,
			Visible:       true,
		}
		if c, ok := cfgMap[m.ID]; ok {
			dto.Visible = c.Visible
			dto.LookbackDaysOverride = c.LookbackDaysOverride
			dto.MaxItemsOverride = c.MaxItemsOverride
			dto.SortOrder = c.SortOrder
		}
		out = append(out, dto)
	}
	return out, nil
}

func (r *NewspaperRepository) GetModuleConfigMap() (map[string]model.NewspaperModuleConfig, error) {
	var cfgRows []model.NewspaperModuleConfig
	if err := r.db.Find(&cfgRows).Error; err != nil {
		return nil, err
	}
	out := make(map[string]model.NewspaperModuleConfig, len(cfgRows))
	for _, c := range cfgRows {
		out[c.ModuleID] = c
	}
	return out, nil
}

func (r *NewspaperRepository) SaveModuleConfigs(items []NewspaperModuleConfigInput) error {
	now := time.Now().UTC()
	return r.db.Transaction(func(tx *gorm.DB) error {
		for _, it := range items {
			if it.ModuleID == "" {
				continue
			}
			lookback := it.LookbackDaysOverride
			if lookback < 0 {
				lookback = 0
			}
			maxItems := it.MaxItemsOverride
			if maxItems < 0 {
				maxItems = 0
			}
			var existing model.NewspaperModuleConfig
			err := tx.Where("module_id = ?", it.ModuleID).First(&existing).Error
			values := map[string]any{
				"visible":                it.Visible,
				"lookback_days_override": lookback,
				"max_items_override":     maxItems,
				"sort_order":             it.SortOrder,
				"updated_at":             now,
			}
			if err != nil {
				if !errors.Is(err, gorm.ErrRecordNotFound) {
					return err
				}
				values["module_id"] = it.ModuleID
				if err := tx.Model(&model.NewspaperModuleConfig{}).Create(values).Error; err != nil {
					return err
				}
			} else {
				if err := tx.Model(&model.NewspaperModuleConfig{}).Where("module_id = ?", it.ModuleID).Updates(values).Error; err != nil {
					return err
				}
			}
		}
		return nil
	})
}
