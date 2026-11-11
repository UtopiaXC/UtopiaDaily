package module

import (
	"context"
	"time"

	"github.com/UtopiaXC/UtopiaDaily/internal/pkg/mailer"
)

const (
	KindScraper = "scraper"
	KindPush    = "push"

	SourceBuiltIn  = "default"
	SourceExternal = "external"
)

var DefaultVersion = "0.2.0"

type Metadata struct {
	ID       string `json:"id"`
	Kind     string `json:"kind"`
	Source   string `json:"source"`
	Version  string `json:"version"`
	Author   string `json:"author"`
	EntryDir string `json:"entry_dir,omitempty"`
}

type ConfigFieldSpec struct {
	Key            string   `json:"key"`
	Type           string   `json:"type"`
	DefaultValue   string   `json:"default_value,omitempty"`
	DefaultItems   []string `json:"default_items,omitempty"`
	Options        []string `json:"options,omitempty"`
	Secret         bool     `json:"secret,omitempty"`
	OverrideSystem bool     `json:"override_system,omitempty"`
	SortOrder      int      `json:"sort_order"`
}

type TaskSpec struct {
	Key            string `json:"key"`
	DefaultCron    string `json:"default_cron"`
	DefaultEnabled bool   `json:"default_enabled"`
}

type ScrapedMedia struct {
	Type string
	URL  string
}

type ScrapedQuotation struct {
	Author  string
	Content string
	URL     string
}

type ScrapedItem struct {
	Title       string
	Summary     string
	Content     string
	ContentType string
	RenderType  string
	SourceName  string
	Author      string
	ExternalURL string
	PublishedAt time.Time
	Fingerprint string
	Media       []ScrapedMedia
	Quotation   *ScrapedQuotation
}

type PushPayload struct {
	Title     string                 `json:"title"`
	Recipient string                 `json:"recipient"`
	Articles  []mailer.DigestArticle `json:"articles"`
}

type TestResult struct {
	Success    bool   `json:"success"`
	MessageKey string `json:"message_key,omitempty"`
	Detail     string `json:"detail,omitempty"`
}

type ConfigAccessor interface {
	GetString(key string) string
	GetInt(key string, fallback int) int
	GetBool(key string) bool
	GetStringSlice(key string) []string
	RawMap() map[string]any
	GetSystemConfig(key string) string
}

type Module interface {
	Meta() Metadata
	ConfigSchema() []ConfigFieldSpec
	Tasks() []TaskSpec
	Locales() map[string]map[string]string
}

type ScraperModule interface {
	Module
	Scrape(ctx context.Context, taskKey string, cfg ConfigAccessor) ([]ScrapedItem, error)
	TestConfig(ctx context.Context, cfg ConfigAccessor) TestResult
}

type PushModule interface {
	Module
	Push(ctx context.Context, taskKey string, cfg ConfigAccessor, payload PushPayload) error
	TestConfig(ctx context.Context, cfg ConfigAccessor) TestResult
}
