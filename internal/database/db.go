package database

import (
	"fmt"
	"strings"
	"time"

	"github.com/UtopiaXC/UtopiaDaily/internal/classification"
	"github.com/UtopiaXC/UtopiaDaily/internal/config"
	"github.com/UtopiaXC/UtopiaDaily/internal/logger"
	"github.com/UtopiaXC/UtopiaDaily/internal/model"
	"github.com/glebarez/sqlite"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func Open(cfg *config.Config) (*gorm.DB, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}
	config.EnsureDirectories(cfg)

	gormCfg := &gorm.Config{
		Logger: logger.NewGormLogger(),
	}

	driver := strings.ToLower(strings.TrimSpace(cfg.Database.Driver))
	if driver == "" {
		driver = "sqlite"
	}

	var db *gorm.DB
	var err error
	switch driver {
	case "sqlite", "sqlite3":
		dsn := cfg.Database.DSN
		if dsn == "" {
			dsn = "data/db/utopia_daily.db"
		}
		db, err = gorm.Open(sqlite.Open(dsn), gormCfg)
		if err == nil {
			sqlDB, sqlErr := db.DB()
			if sqlErr == nil {
				sqlDB.SetMaxOpenConns(1)
			}
			db.Exec("PRAGMA journal_mode = WAL;")
			db.Exec("PRAGMA busy_timeout = 5000;")
			db.Exec("PRAGMA synchronous = NORMAL;")
			db.Exec("PRAGMA foreign_keys = ON;")
		}
	case "postgres", "postgresql":
		db, err = gorm.Open(postgres.Open(cfg.Database.DSN), gormCfg)
	case "mysql":
		db, err = gorm.Open(mysql.Open(cfg.Database.DSN), gormCfg)
	default:
		return nil, fmt.Errorf("unsupported database driver: %s", driver)
	}

	if err != nil {
		return nil, err
	}

	if err := AutoMigrate(db); err != nil {
		return nil, fmt.Errorf("auto migrate failed: %w", err)
	}

	if err := SeedDefaults(db, cfg); err != nil {
		return nil, fmt.Errorf("seed defaults failed: %w", err)
	}

	return db, nil
}

func AutoMigrate(db *gorm.DB) error {
	if db.Migrator().HasTable(&model.Tag{}) && !db.Migrator().HasColumn(&model.Tag{}, "group_id") {
		_ = db.Migrator().DropTable(&model.ScraperResultTag{}, &model.TagKeyword{}, &model.Tag{})
	}
	return db.AutoMigrate(
		&model.Permission{},
		&model.Role{},
		&model.RolePermission{},
		&model.User{},
		&model.UserSession{},
		&model.VerificationToken{},
		&model.SystemConfig{},
		&model.SystemEvent{},
		&model.Module{},
		&model.ModuleConfig{},
		&model.ModuleConfigItem{},
		&model.ScheduleJob{},
		&model.ScraperResult{},
		&model.ScraperResultAttachment{},
		&model.TagProfile{},
		&model.TagGroup{},
		&model.Tag{},
		&model.TagKeyword{},
		&model.ScraperResultTag{},
		&model.ModuleClassificationConfig{},
		&model.ClassificationTask{},
		&model.NewspaperModuleConfig{},
		&model.UserPushSubscription{},
		&model.UserPushSource{},
		&model.PushDeliveryLog{},
	)
}

func SeedDefaults(db *gorm.DB, cfg *config.Config) error {
	now := time.Now()

	createdPerms := make(map[string]bool)
	for _, code := range AllPermissions() {
		var p model.Permission
		if err := db.Where("code = ?", code).First(&p).Error; err != nil {
			p = model.Permission{
				Code:      code,
				GroupCode: PermissionGroup(code),
				CreatedAt: now,
			}
			if err := db.Create(&p).Error; err != nil {
				return err
			}
			createdPerms[code] = true
		}
	}

	var allPerms []model.Permission
	if err := db.Find(&allPerms).Error; err != nil {
		return err
	}
	permByCode := make(map[string]model.Permission, len(allPerms))
	for _, p := range allPerms {
		permByCode[p.Code] = p
	}

	var adminRole model.Role
	if err := db.Where("name = ?", "admin").First(&adminRole).Error; err != nil {
		adminRole = model.Role{
			Name:        "admin",
			Description: "role.admin.desc",
			IsSystem:    true,
			CreatedAt:   now,
			UpdatedAt:   now,
		}
		if err := db.Create(&adminRole).Error; err != nil {
			return err
		}
	}
	if err := db.Model(&adminRole).Association("Permissions").Replace(allPerms); err != nil {
		return err
	}

	var userRole model.Role
	userRoleCreated := false
	if err := db.Where("name = ?", "user").First(&userRole).Error; err != nil {
		userRole = model.Role{
			Name:        "user",
			Description: "role.user.desc",
			IsSystem:    true,
			CreatedAt:   now,
			UpdatedAt:   now,
		}
		if err := db.Create(&userRole).Error; err != nil {
			return err
		}
		userRoleCreated = true
	}
	if userRoleCreated {
		var defPerms []model.Permission
		for _, code := range DefaultUserPermissions() {
			if p, ok := permByCode[code]; ok {
				defPerms = append(defPerms, p)
			}
		}
		if err := db.Model(&userRole).Association("Permissions").Replace(defPerms); err != nil {
			return err
		}
	} else if createdPerms[PermNewspaperView] && userRole.ID > 0 {
		if p, ok := permByCode[PermNewspaperView]; ok {
			var existingCount int64
			_ = db.Table("role_permissions").Where("role_id = ? AND permission_id = ?", userRole.ID, p.ID).Count(&existingCount).Error
			if existingCount == 0 {
				if err := db.Model(&userRole).Association("Permissions").Append(&p); err != nil {
					return err
				}
			}
		}
	}

	for _, spec := range config.SystemFieldRegistry() {
		val := spec.DefaultValue
		if cfg != nil {
			switch spec.Key {
			case "server_name":
				if cfg.Site.Name != "" {
					val = cfg.Site.Name
				}
			case "site_description":
				if cfg.Site.Description != "" {
					val = cfg.Site.Description
				}
			case "base_url":
				if cfg.Site.BaseURL != "" {
					val = cfg.Site.BaseURL
				}
			case "default_locale":
				if cfg.Site.DefaultLocale != "" {
					val = cfg.Site.DefaultLocale
				}
			case "log_level":
				if cfg.Logging.Level != "" {
					val = cfg.Logging.Level
				}
			case "log_permanent_level":
				if cfg.Logging.PermanentLevel != "" {
					val = cfg.Logging.PermanentLevel
				}
			case "smtp_enabled":
				if cfg.SMTP.Enabled {
					val = "true"
				} else {
					val = "false"
				}
			case "smtp_host":
				val = cfg.SMTP.Host
			case "smtp_port":
				if cfg.SMTP.Port > 0 {
					val = fmt.Sprintf("%d", cfg.SMTP.Port)
				}
			case "smtp_encryption":
				if cfg.SMTP.Encryption != "" {
					val = cfg.SMTP.Encryption
				}
			case "smtp_from":
				val = cfg.SMTP.From
			case "smtp_from_name":
				if cfg.SMTP.FromName != "" {
					val = cfg.SMTP.FromName
				}
			case "smtp_user":
				val = cfg.SMTP.User
			case "smtp_password":
				val = cfg.SMTP.Password
			case "newspaper_title":
				if cfg.Site.Name != "" {
					val = cfg.Site.Name
				}
			}
		}

		var existing model.SystemConfig
		if err := db.Where("key = ?", spec.Key).First(&existing).Error; err != nil {
			sc := model.SystemConfig{
				Key:       spec.Key,
				Value:     val,
				UpdatedAt: now,
			}
			if err := db.Create(&sc).Error; err != nil {
				return err
			}
		} else if spec.Key == "system_version" && existing.Value != config.Version {
			db.Model(&existing).Updates(map[string]any{
				"value":      config.Version,
				"updated_at": now,
			})
		}
	}

	var profileCount int64
	if err := db.Model(&model.TagProfile{}).Count(&profileCount).Error; err == nil && profileCount == 0 {
		locale := "en_US"
		var locCfg model.SystemConfig
		if err := db.Where("key = ?", "default_locale").First(&locCfg).Error; err == nil && strings.TrimSpace(locCfg.Value) != "" {
			locale = locCfg.Value
		} else if cfg != nil && strings.TrimSpace(cfg.Site.DefaultLocale) != "" {
			locale = cfg.Site.DefaultLocale
		}
		tpl := classification.DefaultTemplate(locale)
		if err := db.Transaction(func(tx *gorm.DB) error {
			p := model.TagProfile{
				Name:        tpl.Name,
				Description: tpl.Description,
				IsDefault:   true,
				CreatedAt:   now,
				UpdatedAt:   now,
			}
			if err := tx.Create(&p).Error; err != nil {
				return err
			}
			for gIdx, gSpec := range tpl.Groups {
				g := model.TagGroup{
					ProfileID:     p.ID,
					Name:          gSpec.Name,
					Description:   gSpec.Description,
					SelectionMode: gSpec.SelectionMode,
					MaxResults:    gSpec.MaxResults,
					SortOrder:     gIdx,
				}
				if err := tx.Create(&g).Error; err != nil {
					return err
				}
				for tIdx, tSpec := range gSpec.Tags {
					t := model.Tag{
						GroupID:     g.ID,
						Name:        tSpec.Name,
						Description: tSpec.Description,
						Priority:    tSpec.Priority,
						SortOrder:   tIdx,
						CreatedAt:   now,
					}
					if err := tx.Create(&t).Error; err != nil {
						return err
					}
					for _, kw := range tSpec.Keywords {
						trimmed := strings.TrimSpace(kw)
						if trimmed == "" {
							continue
						}
						_ = tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&model.TagKeyword{
							TagID:   t.ID,
							Keyword: trimmed,
						}).Error
					}
				}
			}
			return nil
		}); err != nil {
			return err
		}
	}

	var scraperMods []model.Module
	if err := db.Where("kind = ?", model.ModuleKindScraper).Find(&scraperMods).Error; err == nil {
		for _, m := range scraperMods {
			_ = db.Clauses(clause.OnConflict{DoNothing: true}).Create(&model.NewspaperModuleConfig{
				ModuleID:             m.ID,
				Visible:              true,
				LookbackDaysOverride: 0,
				MaxItemsOverride:     0,
				SortOrder:            0,
				UpdatedAt:            now,
			}).Error
		}
	}

	return nil
}
