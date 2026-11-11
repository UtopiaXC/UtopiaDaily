package external

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/UtopiaXC/UtopiaDaily/internal/module"
	"github.com/UtopiaXC/UtopiaDaily/internal/pkg/runner"
)

type rawScrapedItem struct {
	Title       string `json:"title"`
	Summary     string `json:"summary"`
	Content     string `json:"content"`
	ContentType string `json:"content_type"`
	RenderType  string `json:"render_type"`
	SourceName  string `json:"source_name"`
	Source      string `json:"source"`
	Author      string `json:"author"`
	ExternalURL string `json:"external_url"`
	URL         string `json:"url"`
	FromURL     string `json:"from_url"`
	PublishedAt string `json:"published_at"`
	Timestamp   string `json:"timestamp"`
	Fingerprint string `json:"fingerprint"`
	Media       []struct {
		Type string `json:"type"`
		URL  string `json:"url"`
	} `json:"media"`
	Images    []string `json:"images"`
	Quotation *struct {
		Author  string `json:"author"`
		Content string `json:"content"`
		URL     string `json:"url"`
	} `json:"quotation"`
}

type baseAdapter struct {
	manifest *Manifest
	locales  map[string]map[string]string
	dir      string
	runner   *runner.LocalPluginRunner
}

func (a *baseAdapter) Meta() module.Metadata {
	return module.Metadata{
		ID:       a.manifest.ID,
		Kind:     a.manifest.Kind,
		Source:   module.SourceExternal,
		Version:  a.manifest.Version,
		Author:   a.manifest.Author,
		EntryDir: a.dir,
	}
}

func (a *baseAdapter) ConfigSchema() []module.ConfigFieldSpec {
	return a.manifest.Configs
}

func (a *baseAdapter) Tasks() []module.TaskSpec {
	return a.manifest.Tasks
}

func (a *baseAdapter) Locales() map[string]map[string]string {
	return a.locales
}

func (a *baseAdapter) execute(ctx context.Context, taskID string, cfgMap map[string]any) (*runner.TaskOutput, error) {
	cmd := ResolveCommand(a.dir, a.manifest.Command)
	return a.runner.ExecutePlugin(ctx, a.dir, cmd, runner.TaskInput{
		TaskID: taskID,
		Config: cfgMap,
	})
}

type ScraperAdapter struct {
	baseAdapter
}

func NewScraperAdapter(dir string, m *Manifest, locales map[string]map[string]string, r *runner.LocalPluginRunner) *ScraperAdapter {
	return &ScraperAdapter{
		baseAdapter: baseAdapter{
			manifest: m,
			locales:  locales,
			dir:      dir,
			runner:   r,
		},
	}
}

func (s *ScraperAdapter) Scrape(ctx context.Context, taskKey string, cfg module.ConfigAccessor) ([]module.ScrapedItem, error) {
	out, err := s.execute(ctx, taskKey, cfg.RawMap())
	if err != nil {
		return nil, err
	}
	if !strings.EqualFold(out.Status, "success") && !strings.EqualFold(out.Status, "ok") {
		msg := out.Message
		if msg == "" {
			msg = "external scraper returned failed status"
		}
		return nil, errors.New(msg)
	}
	if len(out.Data) == 0 || string(out.Data) == "null" {
		return nil, nil
	}

	var raws []rawScrapedItem
	if err := json.Unmarshal(out.Data, &raws); err != nil {
		return nil, fmt.Errorf("decode scraped data array: %w", err)
	}

	now := time.Now().UTC()
	items := make([]module.ScrapedItem, 0, len(raws))
	for _, r := range raws {
		extURL := firstNonEmpty(r.ExternalURL, r.URL, r.FromURL)
		src := firstNonEmpty(r.SourceName, r.Source, s.manifest.ID)
		ctype := r.ContentType
		if ctype == "" {
			ctype = "markdown"
		}
		pubTime := parseFlexibleTime(firstNonEmpty(r.PublishedAt, r.Timestamp), now)

		fp := r.Fingerprint
		if fp == "" {
			rawFP := fmt.Sprintf("%s|%s|%s", s.manifest.ID, extURL, r.Title)
			sum := sha256.Sum256([]byte(rawFP))
			fp = hex.EncodeToString(sum[:])
		}

		var media []module.ScrapedMedia
		for _, m := range r.Media {
			if m.URL != "" {
				mt := m.Type
				if mt == "" {
					mt = "image"
				}
				media = append(media, module.ScrapedMedia{Type: mt, URL: m.URL})
			}
		}
		for _, img := range r.Images {
			if img != "" {
				media = append(media, module.ScrapedMedia{Type: "image", URL: img})
			}
		}

		var quot *module.ScrapedQuotation
		if r.Quotation != nil && (r.Quotation.Content != "" || r.Quotation.URL != "") {
			quot = &module.ScrapedQuotation{
				Author:  r.Quotation.Author,
				Content: r.Quotation.Content,
				URL:     r.Quotation.URL,
			}
		}

		summary := r.Summary
		if summary == "" && len(r.Content) > 0 {
			summary = truncateRunes(r.Content, 200)
		}

		renderType := r.RenderType
		if renderType == "" {
			if ctype == "html" {
				renderType = "html"
			} else {
				renderType = "raw"
			}
		}

		items = append(items, module.ScrapedItem{
			Title:       firstNonEmpty(r.Title, src),
			Summary:     summary,
			Content:     r.Content,
			ContentType: ctype,
			RenderType:  renderType,
			SourceName:  src,
			Author:      r.Author,
			ExternalURL: extURL,
			PublishedAt: pubTime,
			Fingerprint: fp,
			Media:       media,
			Quotation:   quot,
		})
	}
	return items, nil
}

func (s *ScraperAdapter) TestConfig(ctx context.Context, cfg module.ConfigAccessor) module.TestResult {
	out, err := s.execute(ctx, "test", cfg.RawMap())
	if err != nil {
		return module.TestResult{
			Success:    false,
			MessageKey: "module.test.failed",
			Detail:     err.Error(),
		}
	}
	if strings.EqualFold(out.Status, "success") || strings.EqualFold(out.Status, "ok") {
		return module.TestResult{
			Success:    true,
			MessageKey: "module.test.passed",
			Detail:     out.Message,
		}
	}
	return module.TestResult{
		Success:    false,
		MessageKey: "module.test.failed",
		Detail:     out.Message,
	}
}

type PushAdapter struct {
	baseAdapter
}

func NewPushAdapter(dir string, m *Manifest, locales map[string]map[string]string, r *runner.LocalPluginRunner) *PushAdapter {
	return &PushAdapter{
		baseAdapter: baseAdapter{
			manifest: m,
			locales:  locales,
			dir:      dir,
			runner:   r,
		},
	}
}

func (p *PushAdapter) Push(ctx context.Context, taskKey string, cfg module.ConfigAccessor, payload module.PushPayload) error {
	cfgMap := make(map[string]any)
	for k, v := range cfg.RawMap() {
		cfgMap[k] = v
	}
	cfgMap["recipient"] = payload.Recipient
	cfgMap["title"] = payload.Title
	cfgMap["articles"] = payload.Articles

	out, err := p.execute(ctx, taskKey, cfgMap)
	if err != nil {
		return err
	}
	if !strings.EqualFold(out.Status, "success") && !strings.EqualFold(out.Status, "ok") {
		if out.Message != "" {
			return errors.New(out.Message)
		}
		return errors.New("external push module failed")
	}
	return nil
}

func (p *PushAdapter) TestConfig(ctx context.Context, cfg module.ConfigAccessor) module.TestResult {
	out, err := p.execute(ctx, "test", cfg.RawMap())
	if err != nil {
		return module.TestResult{
			Success:    false,
			MessageKey: "module.test.failed",
			Detail:     err.Error(),
		}
	}
	if strings.EqualFold(out.Status, "success") || strings.EqualFold(out.Status, "ok") {
		return module.TestResult{
			Success:    true,
			MessageKey: "module.test.passed",
			Detail:     out.Message,
		}
	}
	return module.TestResult{
		Success:    false,
		MessageKey: "module.test.failed",
		Detail:     out.Message,
	}
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

func parseFlexibleTime(s string, fallback time.Time) time.Time {
	s = strings.TrimSpace(s)
	if s == "" {
		return fallback
	}
	formats := []string{
		time.RFC3339,
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05",
		"2006-01-02",
	}
	for _, layout := range formats {
		if t, err := time.Parse(layout, s); err == nil {
			return t.UTC()
		}
	}
	return fallback
}

func truncateRunes(s string, max int) string {
	runes := []rune(s)
	if len(runes) <= max {
		return s
	}
	return string(runes[:max]) + "..."
}
