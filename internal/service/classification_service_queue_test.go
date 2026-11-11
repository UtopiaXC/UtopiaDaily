package service

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/UtopiaXC/UtopiaDaily/internal/model"
	"github.com/UtopiaXC/UtopiaDaily/internal/repository"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func setupClassificationQueueTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Discard,
	})
	if err != nil {
		t.Fatalf("failed opening sqlite memory db: %v", err)
	}
	err = db.AutoMigrate(
		&model.SystemConfig{},
		&model.Module{},
		&model.ScraperResult{},
		&model.TagProfile{},
		&model.TagGroup{},
		&model.Tag{},
		&model.TagKeyword{},
		&model.ScraperResultTag{},
		&model.ModuleClassificationConfig{},
		&model.ClassificationTask{},
	)
	if err != nil {
		t.Fatalf("AutoMigrate failed: %v", err)
	}
	return db
}

func TestClassificationQueueZeroTagsMarksFailedAndTaskDetailShowsTagsAndError(t *testing.T) {
	db := setupClassificationQueueTestDB(t)
	cfgRepo := repository.NewConfigRepository(db)
	clsRepo := repository.NewClassificationRepository(db)

	var mode atomic.Int32 // 0 = low confidence (0 tags), 1 = semantic match
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if mode.Load() == 0 {
			_, _ = w.Write([]byte(`{"message":{"content":"{\"tags\":[{\"group\":\"Topic\",\"tag\":\"Finance\",\"confidence\":0.25,\"reason\":\"weak fit\"}]}"}}`))
			return
		}
		_, _ = w.Write([]byte(`{"message":{"content":"{\"tags\":[{\"group\":\"Topic\",\"tag\":\"Finance\",\"confidence\":0.88,\"reason\":\"Discusses monetary policy and interest rates\"}]}"}}`))
	}))
	defer srv.Close()

	if err := cfgRepo.SetMany(map[string]string{
		"classification_provider":     "llm",
		"classification_confidence":   "60",
		"classification_llm_protocol": "ollama",
		"classification_llm_url":      srv.URL,
		"classification_llm_model":    "qwen3.5:0.8b",
	}); err != nil {
		t.Fatalf("seed system configs failed: %v", err)
	}

	if err := db.Create(&model.Module{
		ID:      "telegram_channel",
		Kind:    model.ModuleKindScraper,
		Source:  model.ModuleSourceBuiltin,
		NameKey: "module.telegram_channel.name",
		Version: "0.2.0",
		Enabled: true,
	}).Error; err != nil {
		t.Fatalf("seed module failed: %v", err)
	}

	_, err := clsRepo.CreateProfile(repository.ProfileSaveInput{
		Name:      "Default Profile",
		IsDefault: true,
		Groups: []repository.ProfileGroupInput{
			{
				Name:          "Topic",
				SelectionMode: "single",
				MaxResults:    1,
				Tags: []repository.ProfileTagInput{
					{
						Name:        "Finance",
						Description: "Markets and macroeconomy",
						Keywords:    []string{"GDP", "Fed"},
					},
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("CreateProfile failed: %v", err)
	}

	now := time.Now().UTC()
	res := model.ScraperResult{
		ModuleID:         "telegram_channel",
		Fingerprint:      "fp-test-1",
		Title:            "Central Bank Policy Update",
		Summary:          "Interest rate guidance shifted overnight.",
		Source:           "FinWire",
		Author:           "Editor",
		FromURL:          "https://example.com/news/1",
		Content:          "Policy makers signaled steadier liquidity conditions for commercial banks.",
		ContentType:      "markdown",
		RenderType:       "markdown",
		DatetimeReleased: now,
		CreatedAt:        now,
	}
	if err := db.Create(&res).Error; err != nil {
		t.Fatalf("create article failed: %v", err)
	}

	added, err := clsRepo.EnqueueResults([]struct {
		ResultID uint
		ModuleID string
	}{{ResultID: res.ID, ModuleID: "telegram_channel"}})
	if err != nil || added != 1 {
		t.Fatalf("EnqueueResults = (%d, %v), want (1, nil)", added, err)
	}

	svc := NewClassificationService(clsRepo, cfgRepo, nil, nil, nil)

	// Run 1: LLM returns only below-threshold candidate -> MUST fail, never complete with 0 tags.
	if !svc.processOne(context.Background()) {
		t.Fatal("expected processOne to process queued task")
	}
	items, _, err := svc.ListQueueTasks("all", 1, 10)
	if err != nil || len(items) != 1 {
		t.Fatalf("ListQueueTasks failed: %v, len=%d", err, len(items))
	}
	task := items[0]
	if task.Status != model.ClassTaskFailed {
		t.Fatalf("expected failed status when 0 tags exceed threshold, got %q (err=%q)", task.Status, task.ErrorMessage)
	}
	if task.AssignedTagsCount != 0 || !strings.Contains(task.ErrorMessage, "confidence threshold") {
		t.Fatalf("expected 0 assigned tags and threshold diagnostic error, got count=%d err=%q", task.AssignedTagsCount, task.ErrorMessage)
	}

	detailFail, err := svc.GetQueueTaskDetail(task.ID)
	if err != nil {
		t.Fatalf("GetQueueTaskDetail (failed) error: %v", err)
	}
	if detailFail.Article == nil || detailFail.Article.Title != "Central Bank Policy Update" {
		t.Fatalf("expected article metadata in detail view, got %+v", detailFail.Article)
	}
	if len(detailFail.AssignedTags) != 0 {
		t.Fatalf("expected 0 assigned tags in detail view on failure, got %d", len(detailFail.AssignedTags))
	}

	// Retry and run 2: LLM classifies semantically without keyword hits -> completed + assigned tags visible in detail.
	mode.Store(1)
	if err := svc.RetryTask(task.ID); err != nil {
		t.Fatalf("RetryTask failed: %v", err)
	}
	if !svc.processOne(context.Background()) {
		t.Fatal("expected processOne to process retried task")
	}

	detailOK, err := svc.GetQueueTaskDetail(task.ID)
	if err != nil {
		t.Fatalf("GetQueueTaskDetail (completed) error: %v", err)
	}
	if detailOK.Status != model.ClassTaskCompleted || detailOK.AssignedTagsCount != 1 {
		t.Fatalf("expected completed task with 1 tag, got status=%s count=%d err=%q", detailOK.Status, detailOK.AssignedTagsCount, detailOK.ErrorMessage)
	}
	if len(detailOK.AssignedTags) != 1 {
		t.Fatalf("expected 1 assigned tag detail, got %+v", detailOK.AssignedTags)
	}
	gotTag := detailOK.AssignedTags[0]
	if gotTag.GroupName != "Topic" || gotTag.TagName != "Finance" || gotTag.Provider != "llm" || gotTag.Confidence != 0.88 {
		t.Fatalf("unexpected assigned tag detail: %+v", gotTag)
	}

	if _, missingErr := svc.GetQueueTaskDetail(999999); missingErr == nil {
		t.Fatal("expected 404 error for nonexistent queue task detail")
	}
}

func TestClassificationTextLimitConfigAndModuleOverride(t *testing.T) {
	db := setupClassificationQueueTestDB(t)
	cfgRepo := repository.NewConfigRepository(db)
	clsRepo := repository.NewClassificationRepository(db)
	modRepo := repository.NewModuleRepository(db)
	cfgSvc := NewConfigService(cfgRepo, nil, nil, nil, nil, nil, nil)
	clsSvc := NewClassificationService(clsRepo, cfgRepo, modRepo, nil, nil)

	if err := db.Create(&model.Module{
		ID:      "telegram_channel",
		Kind:    model.ModuleKindScraper,
		Source:  model.ModuleSourceBuiltin,
		NameKey: "module.telegram_channel.name",
		Version: "0.2.0",
		Enabled: true,
	}).Error; err != nil {
		t.Fatalf("seed module failed: %v", err)
	}

	// Default fallback when config keys not yet set
	optsDefault, err := clsSvc.resolveEffectiveOptions("telegram_channel")
	if err != nil {
		t.Fatalf("resolveEffectiveOptions default failed: %v", err)
	}
	if optsDefault.MaxInputChars != 2000 || optsDefault.TruncateStrategy != "title_head_tail" {
		t.Fatalf("expected default limit=2000 strategy=title_head_tail, got limit=%d strategy=%q", optsDefault.MaxInputChars, optsDefault.TruncateStrategy)
	}

	// Validate ConfigService accepts 0 (unlimited) and rejects <0 or invalid strategy
	if err := cfgSvc.UpdateConfig("classification_max_input_chars", "-1", false); err == nil {
		t.Fatal("expected negative classification_max_input_chars to fail validation")
	}
	if err := cfgSvc.UpdateConfig("classification_truncate_strategy", "middle_only", false); err == nil {
		t.Fatal("expected unknown classification_truncate_strategy to fail validation")
	}
	if err := cfgSvc.UpdateBatchConfigs(map[string]any{
		"classification_max_input_chars":   "0",
		"classification_truncate_strategy": "tail",
	}, false); err != nil {
		t.Fatalf("UpdateBatchConfigs valid text settings failed: %v", err)
	}

	optsUnlimited, err := clsSvc.resolveEffectiveOptions("telegram_channel")
	if err != nil {
		t.Fatalf("resolveEffectiveOptions system unlimited failed: %v", err)
	}
	if optsUnlimited.MaxInputChars != 0 || optsUnlimited.TruncateStrategy != "tail" {
		t.Fatalf("expected system unlimited limit=0 strategy=tail, got limit=%d strategy=%q", optsUnlimited.MaxInputChars, optsUnlimited.TruncateStrategy)
	}

	// Module override with nil limit & empty strategy inherits system (0, tail)
	if err := clsSvc.SaveModuleConfig("telegram_channel", ModuleClassificationConfigView{
		OverrideSystem:   true,
		Provider:         "llm",
		Confidence:       65,
		MaxInputChars:    nil,
		TruncateStrategy: "",
	}); err != nil {
		t.Fatalf("SaveModuleConfig inherit failed: %v", err)
	}
	optsInherit, err := clsSvc.resolveEffectiveOptions("telegram_channel")
	if err != nil {
		t.Fatalf("resolveEffectiveOptions module inherit failed: %v", err)
	}
	if optsInherit.MaxInputChars != 0 || optsInherit.TruncateStrategy != "tail" {
		t.Fatalf("expected module inherit limit=0 strategy=tail, got limit=%d strategy=%q", optsInherit.MaxInputChars, optsInherit.TruncateStrategy)
	}

	// Module explicit override to 800 & head, and test 0 explicit override over system 3500
	if err := cfgSvc.UpdateConfig("classification_max_input_chars", "3500", false); err != nil {
		t.Fatalf("set system limit=3500 failed: %v", err)
	}
	limit800 := 800
	if err := clsSvc.SaveModuleConfig("telegram_channel", ModuleClassificationConfigView{
		OverrideSystem:   true,
		Provider:         "onnx",
		Confidence:       70,
		MaxInputChars:    &limit800,
		TruncateStrategy: "head",
	}); err != nil {
		t.Fatalf("SaveModuleConfig 800/head failed: %v", err)
	}
	opts800, err := clsSvc.resolveEffectiveOptions("telegram_channel")
	if err != nil {
		t.Fatalf("resolveEffectiveOptions module 800 failed: %v", err)
	}
	if opts800.MaxInputChars != 800 || opts800.TruncateStrategy != "head" {
		t.Fatalf("expected module override limit=800 strategy=head, got limit=%d strategy=%q", opts800.MaxInputChars, opts800.TruncateStrategy)
	}

	limitZero := 0
	if err := clsSvc.SaveModuleConfig("telegram_channel", ModuleClassificationConfigView{
		OverrideSystem:   true,
		Provider:         "llm",
		Confidence:       60,
		MaxInputChars:    &limitZero,
		TruncateStrategy: "title_head_tail",
	}); err != nil {
		t.Fatalf("SaveModuleConfig 0/title_head_tail failed: %v", err)
	}
	view, err := clsSvc.GetModuleConfig("telegram_channel")
	if err != nil {
		t.Fatalf("GetModuleConfig failed: %v", err)
	}
	if view.MaxInputChars == nil || *view.MaxInputChars != 0 || view.TruncateStrategy != "title_head_tail" {
		t.Fatalf("expected persisted zero override, got limit=%v strategy=%q", view.MaxInputChars, view.TruncateStrategy)
	}
	optsZero, err := clsSvc.resolveEffectiveOptions("telegram_channel")
	if err != nil {
		t.Fatalf("resolveEffectiveOptions module 0 failed: %v", err)
	}
	if optsZero.MaxInputChars != 0 || optsZero.TruncateStrategy != "title_head_tail" {
		t.Fatalf("expected module zero override over 3500, got limit=%d strategy=%q", optsZero.MaxInputChars, optsZero.TruncateStrategy)
	}
}
