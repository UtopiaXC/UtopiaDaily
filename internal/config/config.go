package config

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

var Version = "0.2.0"

type Config struct {
	Site     SiteConfig     `yaml:"site"`
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	Storage  StorageConfig  `yaml:"storage"`
	Security SecurityConfig `yaml:"security"`
	SMTP     SMTPConfig     `yaml:"smtp"`
	Logging  LoggingConfig  `yaml:"logging"`
}

type SiteConfig struct {
	Name          string `yaml:"name"`
	Description   string `yaml:"description"`
	BaseURL       string `yaml:"base_url"`
	DefaultLocale string `yaml:"default_locale"`
}

type ServerConfig struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
	Mode string `yaml:"mode"`
}

type DatabaseConfig struct {
	Driver string `yaml:"driver"`
	DSN    string `yaml:"dsn"`
}

type StorageConfig struct {
	CacheDir   string `yaml:"cache_dir"`
	PluginsDir string `yaml:"plugins_dir"`
}

type SecurityConfig struct {
	JWTSecret         string `yaml:"jwt_secret"`
	JWTAccessMinutes  int    `yaml:"jwt_access_minutes"`
	JWTRefreshDays    int    `yaml:"jwt_refresh_days"`
	PoWDifficulty     int64  `yaml:"pow_difficulty"`
	MaxLoginAttempts  int    `yaml:"max_login_attempts"`
	LockoutMinutes    int    `yaml:"lockout_minutes"`
	AllowRegistration bool   `yaml:"allow_registration"`
}

type SMTPConfig struct {
	Enabled    bool   `yaml:"enabled"`
	Host       string `yaml:"host"`
	Port       int    `yaml:"port"`
	Encryption string `yaml:"encryption"`
	From       string `yaml:"from"`
	FromName   string `yaml:"from_name"`
	User       string `yaml:"user"`
	Password   string `yaml:"password"`
}

type LoggingConfig struct {
	Level          string `yaml:"level"`
	Dir            string `yaml:"dir"`
	MaxSizeMB      int    `yaml:"max_size_mb"`
	MaxBackups     int    `yaml:"max_backups"`
	MaxAgeDays     int    `yaml:"max_age_days"`
	Compress       bool   `yaml:"compress"`
	PermanentLevel string `yaml:"permanent_level"`
}

func DefaultConfig() *Config {
	return &Config{
		Site: SiteConfig{
			Name:          "Utopia Daily",
			Description:   "Modular news aggregation and distribution platform",
			BaseURL:       "http://localhost:8000",
			DefaultLocale: "en_US",
		},
		Server: ServerConfig{
			Host: "0.0.0.0",
			Port: 8000,
			Mode: "release",
		},
		Database: DatabaseConfig{
			Driver: "sqlite",
			DSN:    "data/db/utopia_daily.db",
		},
		Storage: StorageConfig{
			CacheDir:   "cache",
			PluginsDir: "plugins",
		},
		Security: SecurityConfig{
			JWTSecret:         "",
			JWTAccessMinutes:  120,
			JWTRefreshDays:    30,
			PoWDifficulty:     50000,
			MaxLoginAttempts:  5,
			LockoutMinutes:    15,
			AllowRegistration: false,
		},
		SMTP: SMTPConfig{
			Enabled:    false,
			Host:       "",
			Port:       465,
			Encryption: "ssl",
			From:       "",
			FromName:   "Utopia Daily",
			User:       "",
			Password:   "",
		},
		Logging: LoggingConfig{
			Level:          "info",
			Dir:            "logs",
			MaxSizeMB:      20,
			MaxBackups:     10,
			MaxAgeDays:     30,
			Compress:       true,
			PermanentLevel: "error",
		},
	}
}

func LoadConfig(path string) (*Config, error) {
	cfg := DefaultConfig()

	if path != "" {
		data, err := os.ReadFile(path)
		if err == nil {
			if err := yaml.Unmarshal(data, cfg); err != nil {
				return nil, fmt.Errorf("failed to parse config file: %w", err)
			}
		} else if !os.IsNotExist(err) {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	applyEnvOverrides(cfg)

	configChanged := false
	if isWeakSecret(cfg.Security.JWTSecret) {
		buf := make([]byte, 32)
		if _, err := rand.Read(buf); err == nil {
			cfg.Security.JWTSecret = hex.EncodeToString(buf)
			configChanged = true
		}
	}

	if path != "" {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			configChanged = true
		}
		if configChanged {
			if err := SaveConfig(path, cfg); err != nil {
				slog.Error("Failed to persist config", "path", path, "error", err)
			}
		}
	}

	return cfg, nil
}

func SaveConfig(path string, cfg *Config) error {
	dir := filepath.Dir(path)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0700); err != nil {
			return fmt.Errorf("failed to create config dir: %w", err)
		}
	}
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal yaml: %w", err)
	}
	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}
	return nil
}

func EnsureDirectories(cfg *Config) {
	dirs := []string{
		"data",
		"data/config",
		"data/db",
		"cache",
		"cache/locales",
		"cache/onnx_runtime",
		"logs",
		"models",
		"models/onnx",
		"plugins",
		"plugins/scrapers",
		"plugins/push",
	}
	if cfg != nil {
		if cfg.Database.Driver == "sqlite" || cfg.Database.Driver == "sqlite3" || cfg.Database.Driver == "" {
			if dir := filepath.Dir(cfg.Database.DSN); dir != "" && dir != "." {
				dirs = append(dirs, dir)
			}
		}
		if cfg.Storage.CacheDir != "" {
			dirs = append(dirs,
				cfg.Storage.CacheDir,
				filepath.Join(cfg.Storage.CacheDir, "locales"),
				filepath.Join(cfg.Storage.CacheDir, "onnx_runtime"),
			)
		}
		if cfg.Storage.PluginsDir != "" {
			dirs = append(dirs,
				cfg.Storage.PluginsDir,
				filepath.Join(cfg.Storage.PluginsDir, "scrapers"),
				filepath.Join(cfg.Storage.PluginsDir, "push"),
			)
		}
		if cfg.Logging.Dir != "" {
			dirs = append(dirs, cfg.Logging.Dir)
		}
	}
	for _, d := range dirs {
		perm := os.FileMode(0755)
		if strings.HasPrefix(d, "data/config") || strings.HasPrefix(d, "data/db") {
			perm = 0700
		}
		_ = os.MkdirAll(d, perm)
	}
}

func isWeakSecret(s string) bool {
	trimmed := strings.TrimSpace(s)
	return len(trimmed) < 16 || trimmed == "CHANGE_ME"
}

func applyEnvOverrides(cfg *Config) {
	if v := os.Getenv("UDAILY_SERVER_HOST"); v != "" {
		cfg.Server.Host = strings.TrimSpace(v)
	}
	if v := os.Getenv("UDAILY_SERVER_PORT"); v != "" {
		if p, err := strconv.Atoi(v); err == nil && p > 0 {
			cfg.Server.Port = p
		}
	}
	if v := os.Getenv("UDAILY_SERVER_MODE"); v != "" {
		cfg.Server.Mode = strings.TrimSpace(v)
	}
	if v := os.Getenv("UDAILY_BASE_URL"); v != "" {
		cfg.Site.BaseURL = strings.TrimRight(strings.TrimSpace(v), "/")
	}
	if v := os.Getenv("UDAILY_DB_DRIVER"); v != "" {
		cfg.Database.Driver = strings.ToLower(strings.TrimSpace(v))
	}
	if v := os.Getenv("UDAILY_DB_DSN"); v != "" {
		cfg.Database.DSN = strings.TrimSpace(v)
	}
	if v := os.Getenv("UDAILY_JWT_SECRET"); v != "" {
		cfg.Security.JWTSecret = strings.TrimSpace(v)
	} else if v := os.Getenv("JWT_SECRET"); v != "" {
		cfg.Security.JWTSecret = strings.TrimSpace(v)
	}
	if v := os.Getenv("UDAILY_LOG_LEVEL"); v != "" {
		cfg.Logging.Level = strings.TrimSpace(v)
	}
	if v := os.Getenv("UDAILY_PLUGINS_DIR"); v != "" {
		cfg.Storage.PluginsDir = strings.TrimSpace(v)
	}
	if v := os.Getenv("UDAILY_CACHE_DIR"); v != "" {
		cfg.Storage.CacheDir = strings.TrimSpace(v)
	}
}
