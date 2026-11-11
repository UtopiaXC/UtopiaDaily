package repository

import (
	"time"

	"github.com/UtopiaXC/UtopiaDaily/internal/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type PushRepository struct {
	db *gorm.DB
}

func NewPushRepository(db *gorm.DB) *PushRepository {
	return &PushRepository{db: db}
}

func (r *PushRepository) ListByUser(userID uint) ([]model.UserPushSubscription, error) {
	var subs []model.UserPushSubscription
	err := r.db.Preload("Sources").Where("user_id = ?", userID).Order("id ASC").Find(&subs).Error
	return subs, err
}

func (r *PushRepository) GetByID(id uint) (*model.UserPushSubscription, error) {
	var sub model.UserPushSubscription
	if err := r.db.Preload("Sources").Where("id = ?", id).First(&sub).Error; err != nil {
		return nil, err
	}
	return &sub, nil
}

func (r *PushRepository) GetByUserAndModule(userID uint, pushModuleID string) (*model.UserPushSubscription, error) {
	var sub model.UserPushSubscription
	if err := r.db.Preload("Sources").Where("user_id = ? AND push_module_id = ?", userID, pushModuleID).First(&sub).Error; err != nil {
		return nil, err
	}
	return &sub, nil
}

func (r *PushRepository) UpsertSubscription(sub *model.UserPushSubscription, sourceModuleIDs []string) error {
	now := time.Now().UTC()
	return r.db.Transaction(func(tx *gorm.DB) error {
		var existing model.UserPushSubscription
		err := tx.Where("user_id = ? AND push_module_id = ?", sub.UserID, sub.PushModuleID).First(&existing).Error
		if err != nil {
			sub.CreatedAt = now
			sub.UpdatedAt = now
			if err := tx.Create(sub).Error; err != nil {
				return err
			}
		} else {
			sub.ID = existing.ID
			sub.CreatedAt = existing.CreatedAt
			sub.UpdatedAt = now
			if err := tx.Model(&existing).Updates(map[string]any{
				"enabled":        sub.Enabled,
				"target":         sub.Target,
				"cron_expr":      sub.CronExpr,
				"digest_title":   sub.DigestTitle,
				"max_items":      sub.MaxItems,
				"lookback_hours": sub.LookbackHours,
				"updated_at":     now,
			}).Error; err != nil {
				return err
			}
		}

		if err := tx.Where("subscription_id = ?", sub.ID).Delete(&model.UserPushSource{}).Error; err != nil {
			return err
		}
		for _, modID := range sourceModuleIDs {
			if modID == "" {
				continue
			}
			src := model.UserPushSource{
				SubscriptionID:  sub.ID,
				ScraperModuleID: modID,
			}
			if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&src).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *PushRepository) DeleteSubscription(userID, subID uint) error {
	return r.db.Where("id = ? AND user_id = ?", subID, userID).Delete(&model.UserPushSubscription{}).Error
}

func (r *PushRepository) ListEnabledSubscriptions() ([]model.UserPushSubscription, error) {
	var subs []model.UserPushSubscription
	err := r.db.Preload("Sources").Preload("User").Where("enabled = ?", true).Order("id ASC").Find(&subs).Error
	return subs, err
}

func (r *PushRepository) MarkPushed(subID uint) error {
	now := time.Now().UTC()
	return r.db.Model(&model.UserPushSubscription{}).Where("id = ?", subID).Updates(map[string]any{
		"last_pushed_at": now,
		"updated_at":     now,
	}).Error
}

func (r *PushRepository) CreateDeliveryLog(entry *model.PushDeliveryLog) error {
	if entry.CreatedAt.IsZero() {
		entry.CreatedAt = time.Now().UTC()
	}
	return r.db.Create(entry).Error
}

func (r *PushRepository) ListDeliveryLogs(limit int) ([]model.PushDeliveryLog, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	var logs []model.PushDeliveryLog
	err := r.db.Order("created_at DESC, id DESC").Limit(limit).Find(&logs).Error
	return logs, err
}
