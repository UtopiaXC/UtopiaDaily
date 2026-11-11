package repository

import (
	"time"

	"github.com/UtopiaXC/UtopiaDaily/internal/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ConfigRepository struct {
	db *gorm.DB
}

func NewConfigRepository(db *gorm.DB) *ConfigRepository {
	return &ConfigRepository{db: db}
}

func (r *ConfigRepository) GetMap() (map[string]string, error) {
	var list []model.SystemConfig
	if err := r.db.Find(&list).Error; err != nil {
		return nil, err
	}
	res := make(map[string]string, len(list))
	for _, item := range list {
		res[item.Key] = item.Value
	}
	return res, nil
}

func (r *ConfigRepository) Get(key string) (string, error) {
	var cfg model.SystemConfig
	if err := r.db.Where("key = ?", key).First(&cfg).Error; err != nil {
		return "", err
	}
	return cfg.Value, nil
}

func (r *ConfigRepository) Set(key, value string) error {
	now := time.Now().UTC()
	return r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "key"}},
		DoUpdates: clause.AssignmentColumns([]string{"value", "updated_at"}),
	}).Create(&model.SystemConfig{
		Key:       key,
		Value:     value,
		UpdatedAt: now,
	}).Error
}

func (r *ConfigRepository) SetMany(values map[string]string) error {
	now := time.Now().UTC()
	return r.db.Transaction(func(tx *gorm.DB) error {
		for key, val := range values {
			if err := tx.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "key"}},
				DoUpdates: clause.AssignmentColumns([]string{"value", "updated_at"}),
			}).Create(&model.SystemConfig{
				Key:       key,
				Value:     val,
				UpdatedAt: now,
			}).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
