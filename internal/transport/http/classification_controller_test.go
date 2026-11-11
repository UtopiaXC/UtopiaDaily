package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/UtopiaXC/UtopiaDaily/internal/model"
	"github.com/UtopiaXC/UtopiaDaily/internal/repository"
	"github.com/UtopiaXC/UtopiaDaily/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestClassificationControllerGetQueueTaskDetail(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Discard})
	if err != nil {
		t.Fatalf("open sqlite memory db: %v", err)
	}
	if err := db.AutoMigrate(
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
	); err != nil {
		t.Fatalf("AutoMigrate: %v", err)
	}

	cfgRepo := repository.NewConfigRepository(db)
	clsRepo := repository.NewClassificationRepository(db)
	clsSvc := service.NewClassificationService(clsRepo, cfgRepo, nil, nil, nil)
	ctl := NewClassificationController(clsSvc)

	now := time.Now().UTC()
	article := model.ScraperResult{
		ModuleID:         "telegram_channel",
		Fingerprint:      "fp-http-1",
		Title:            "Morning Brief",
		Summary:          "Key headlines",
		Source:           "ChannelA",
		Author:           "Editor",
		FromURL:          "https://t.me/channel_a/10",
		Content:          "Policy [announcement](https://example.com) details.",
		ContentType:      "markdown",
		RenderType:       "markdown",
		DatetimeReleased: now,
		CreatedAt:        now,
	}
	if err := db.Create(&article).Error; err != nil {
		t.Fatalf("create article: %v", err)
	}

	grp := model.TagGroup{ProfileID: 1, Name: "Topic", SelectionMode: "single", MaxResults: 1}
	if err := db.Create(&grp).Error; err != nil {
		t.Fatalf("create group: %v", err)
	}
	tag := model.Tag{GroupID: grp.ID, Name: "Finance"}
	if err := db.Create(&tag).Error; err != nil {
		t.Fatalf("create tag: %v", err)
	}
	if err := db.Create(&model.ScraperResultTag{
		ResultID:   article.ID,
		TagID:      tag.ID,
		Confidence: 0.91,
		Provider:   "llm",
		Reason:     "Matches financial policy topic",
		CreatedAt:  now,
	}).Error; err != nil {
		t.Fatalf("create result tag: %v", err)
	}

	task := model.ClassificationTask{
		ResultID:          article.ID,
		ModuleID:          "telegram_channel",
		Status:            model.ClassTaskCompleted,
		Provider:          "llm",
		ModelName:         "qwen3.5:0.8b",
		Attempts:          1,
		AssignedTagsCount: 1,
		CreatedAt:         now,
		FinishedAt:        &now,
	}
	if err := db.Create(&task).Error; err != nil {
		t.Fatalf("create task: %v", err)
	}

	r := gin.New()
	r.GET("/api/dashboard/classification/queue/tasks/:id", ctl.GetQueueTaskDetail)

	t.Run("valid task id returns full detail", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/dashboard/classification/queue/tasks/%d", task.ID), nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
		}

		var got repository.ClassificationTaskDetailView
		if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
			t.Fatalf("decode response: %v", err)
		}
		if got.ID != task.ID || got.Article == nil || got.Article.Title != "Morning Brief" {
			t.Fatalf("unexpected task detail payload: %+v", got)
		}
		if len(got.AssignedTags) != 1 || got.AssignedTags[0].TagName != "Finance" || got.AssignedTags[0].GroupName != "Topic" {
			t.Fatalf("unexpected assigned tags: %+v", got.AssignedTags)
		}
	})

	t.Run("invalid task id returns 400", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/dashboard/classification/queue/tasks/abc", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", w.Code)
		}
	})

	t.Run("nonexistent task id returns 404", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/dashboard/classification/queue/tasks/999999", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", w.Code)
		}
	})
}
