package event

import (
	"context"
	"encoding/json"
	"log/slog"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/UtopiaXC/UtopiaDaily/internal/model"
	"github.com/UtopiaXC/UtopiaDaily/internal/repository"
)

type Event struct {
	Level      string
	Category   string
	EventKey   string
	Params     map[string]any
	Details    map[string]any
	SourceType string
	SourceID   string
}

type Handler func(ctx context.Context, ev Event)

type Bus struct {
	mu       sync.RWMutex
	handlers []Handler
}

func NewBus() *Bus {
	return &Bus{}
}

func (b *Bus) Subscribe(h Handler) {
	if h == nil {
		return
	}
	b.mu.Lock()
	b.handlers = append(b.handlers, h)
	b.mu.Unlock()
}

func (b *Bus) Publish(ctx context.Context, ev Event) {
	if ev.Level == "" {
		ev.Level = model.EventLevelNormal
	}
	if ev.Category == "" {
		ev.Category = model.EventCategorySystem
	}

	b.mu.RLock()
	handlers := make([]Handler, len(b.handlers))
	copy(handlers, b.handlers)
	b.mu.RUnlock()

	for _, h := range handlers {
		h(ctx, ev)
	}
}

func (b *Bus) Emit(level, category, eventKey string, params, details map[string]any, sourceType, sourceID string) {
	b.Publish(context.Background(), Event{
		Level:      level,
		Category:   category,
		EventKey:   eventKey,
		Params:     params,
		Details:    details,
		SourceType: sourceType,
		SourceID:   sourceID,
	})
}

type DatabaseRecorder struct {
	eventRepo  *repository.EventRepository
	configRepo *repository.ConfigRepository
	counter    atomic.Int64
}

func NewDatabaseRecorder(eventRepo *repository.EventRepository, configRepo *repository.ConfigRepository) *DatabaseRecorder {
	return &DatabaseRecorder{
		eventRepo:  eventRepo,
		configRepo: configRepo,
	}
}

func (r *DatabaseRecorder) Handle(_ context.Context, ev Event) {
	var paramsJSON, detailsJSON string
	if len(ev.Params) > 0 {
		if raw, err := json.Marshal(ev.Params); err == nil {
			paramsJSON = string(raw)
		}
	}
	if len(ev.Details) > 0 {
		if raw, err := json.Marshal(ev.Details); err == nil {
			detailsJSON = string(raw)
		}
	}

	rec := &model.SystemEvent{
		Level:       ev.Level,
		Category:    ev.Category,
		EventKey:    ev.EventKey,
		ParamsJSON:  paramsJSON,
		DetailsJSON: detailsJSON,
		SourceType:  ev.SourceType,
		SourceID:    ev.SourceID,
		CreatedAt:   time.Now().UTC(),
	}
	if err := r.eventRepo.Create(rec); err != nil {
		slog.Error("Persist system event failed", "event_key", ev.EventKey, "error", err)
		return
	}

	if r.counter.Add(1)%25 == 0 {
		go r.prune()
	}
}

func (r *DatabaseRecorder) prune() {
	maxCount := 1000
	maxDays := 30
	if r.configRepo != nil {
		if v, err := r.configRepo.Get("event_max_count"); err == nil {
			if n, err := strconv.Atoi(v); err == nil && n > 0 {
				maxCount = n
			}
		}
		if v, err := r.configRepo.Get("event_max_days"); err == nil {
			if n, err := strconv.Atoi(v); err == nil && n > 0 {
				maxDays = n
			}
		}
	}
	_ = r.eventRepo.Prune(maxCount, maxDays)
}
