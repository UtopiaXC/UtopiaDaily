package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/UtopiaXC/UtopiaDaily/internal/classification"
	"github.com/UtopiaXC/UtopiaDaily/internal/common"
	"github.com/UtopiaXC/UtopiaDaily/internal/config"
	"github.com/UtopiaXC/UtopiaDaily/internal/event"
	"github.com/UtopiaXC/UtopiaDaily/internal/model"
	"github.com/UtopiaXC/UtopiaDaily/internal/repository"
	"gorm.io/gorm"
)

type QueueStatusView struct {
	WorkerState   string                `json:"worker_state"`
	Paused        bool                  `json:"paused"`
	CurrentTaskID uint64                `json:"current_task_id"`
	Stats         repository.QueueStats `json:"stats"`
}

type ModuleClassificationConfigView struct {
	ModuleID         string `json:"module_id"`
	OverrideSystem   bool   `json:"override_system"`
	Provider         string `json:"provider"`
	ModelName        string `json:"model_name"`
	Confidence       int    `json:"confidence"`
	MaxInputChars    *int   `json:"max_input_chars"`
	TruncateStrategy string `json:"truncate_strategy"`
	ProfileID        *uint  `json:"profile_id"`
	LLMProtocol      string `json:"llm_protocol"`
	LLMURL           string `json:"llm_url"`
	LLMAPIKey        string `json:"llm_api_key"`
	LLMModel         string `json:"llm_model"`
	LLMPrompt        string `json:"llm_prompt"`
}

const (
	ClassJobONNXEnv      = "onnx_env"
	ClassJobONNXDownload = "onnx_download"
	ClassJobOllamaPull   = "ollama_pull"

	ClassJobRunning   = "running"
	ClassJobCompleted = "completed"
	ClassJobFailed    = "failed"
	ClassJobCancelled = "cancelled"
)

type ClassificationJob struct {
	ID             string     `json:"id"`
	Kind           string     `json:"kind"`
	Target         string     `json:"target"`
	EndpointURL    string     `json:"endpoint_url,omitempty"`
	Status         string     `json:"status"`
	Stage          string     `json:"stage"`
	Message        string     `json:"message,omitempty"`
	CurrentItem    string     `json:"current_item,omitempty"`
	CompletedBytes int64      `json:"completed_bytes"`
	TotalBytes     int64      `json:"total_bytes"`
	Percent        float64    `json:"percent"`
	Indeterminate  bool       `json:"indeterminate"`
	Output         string     `json:"output,omitempty"`
	Error          string     `json:"error,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
	FinishedAt     *time.Time `json:"finished_at,omitempty"`
}

type classJobEntry struct {
	job    ClassificationJob
	cancel context.CancelFunc
}

type ClassificationOverviewView struct {
	SystemProvider   string                            `json:"system_provider"`
	SystemConfidence int                               `json:"system_confidence"`
	ONNXEnv          classification.ONNXEnvStatus      `json:"onnx_env"`
	ONNXPresetGroups []classification.PresetModelGroup `json:"onnx_presets"`
	ONNXLocalModels  []classification.LocalONNXModel   `json:"onnx_local_models"`
	LLMPresetModels  []string                          `json:"llm_presets"`
	Queue            QueueStatusView                   `json:"queue"`
	Jobs             []ClassificationJob               `json:"jobs"`
}

type ClassificationService struct {
	clsRepo *repository.ClassificationRepository
	cfgRepo *repository.ConfigRepository
	modRepo *repository.ModuleRepository
	onnxMgr *classification.ONNXManager
	llmCls  *classification.LLMClassifier
	bus     *event.Bus
	cfg     *config.Config

	bgCtx     context.Context
	bgCancel  context.CancelFunc
	startOnce sync.Once
	stopOnce  sync.Once
	doneCh    chan struct{}
	wakeCh    chan struct{}
	paused    atomic.Bool
	runningID atomic.Uint64

	jobsMu   sync.RWMutex
	jobs     map[string]*classJobEntry
	jobOrder []string
	jobSeq   atomic.Uint64
	jobsWG   sync.WaitGroup
}

func NewClassificationService(
	clsRepo *repository.ClassificationRepository,
	cfgRepo *repository.ConfigRepository,
	modRepo *repository.ModuleRepository,
	bus *event.Bus,
	cfg *config.Config,
) *ClassificationService {
	cacheDir := "cache"
	if cfg != nil && cfg.Storage.CacheDir != "" {
		cacheDir = cfg.Storage.CacheDir
	}
	bgCtx, bgCancel := context.WithCancel(context.Background())
	return &ClassificationService{
		clsRepo:  clsRepo,
		cfgRepo:  cfgRepo,
		modRepo:  modRepo,
		onnxMgr:  classification.NewONNXManager(cacheDir, "models/onnx"),
		llmCls:   classification.NewLLMClassifier(),
		bus:      bus,
		cfg:      cfg,
		bgCtx:    bgCtx,
		bgCancel: bgCancel,
		wakeCh:   make(chan struct{}, 1),
		jobs:     make(map[string]*classJobEntry),
	}
}

func (s *ClassificationService) Start() {
	s.startOnce.Do(func() {
		_ = s.onnxMgr.EnsureAssets()
		_ = s.clsRepo.ResetStaleRunning()

		if s.bgCtx == nil {
			s.bgCtx, s.bgCancel = context.WithCancel(context.Background())
		}
		s.doneCh = make(chan struct{})

		go s.workerLoop(s.bgCtx)
	})
}

func (s *ClassificationService) Stop() {
	s.stopOnce.Do(func() {
		if s.bgCancel != nil {
			s.bgCancel()
		}
		if s.doneCh != nil {
			<-s.doneCh
		}
		s.jobsWG.Wait()
	})
}

func (s *ClassificationService) Wake() {
	select {
	case s.wakeCh <- struct{}{}:
	default:
	}
}

func (s *ClassificationService) PauseWorker() {
	s.paused.Store(true)
}

func (s *ClassificationService) ResumeWorker() {
	s.paused.Store(false)
	s.Wake()
}

func (s *ClassificationService) WorkerState() (string, uint64) {
	if s.paused.Load() {
		return "paused", s.runningID.Load()
	}
	if cur := s.runningID.Load(); cur > 0 {
		return "running", cur
	}
	return "idle", 0
}

func (s *ClassificationService) workerLoop(ctx context.Context) {
	defer close(s.doneCh)
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		if ctx.Err() != nil {
			return
		}
		if !s.paused.Load() {
			for s.processOne(ctx) {
				if ctx.Err() != nil || s.paused.Load() {
					break
				}
			}
		}

		select {
		case <-ctx.Done():
			return
		case <-s.wakeCh:
		case <-ticker.C:
		}
	}
}

func (s *ClassificationService) processOne(ctx context.Context) bool {
	task, err := s.clsRepo.ClaimNextPending()
	if err != nil || task == nil {
		return false
	}

	s.runningID.Store(uint64(task.ID))
	defer s.runningID.Store(0)

	if task.Result == nil {
		_ = s.clsRepo.SkipTask(task.ID, "article missing")
		return true
	}

	opts, err := s.resolveEffectiveOptions(task.ModuleID)
	if err != nil {
		_ = s.clsRepo.FailTask(task.ID, opts.Provider, opts.ModelName, err.Error())
		return true
	}

	if opts.Provider == "" || opts.Provider == "none" {
		_ = s.clsRepo.SkipTask(task.ID, "classification provider disabled")
		return true
	}

	item := classification.InputItem{
		ResultID:    task.Result.ID,
		Fingerprint: task.Result.Fingerprint,
		Title:       task.Result.Title,
		Summary:     task.Result.Summary,
		Content:     task.Result.Content,
		SourceName:  task.Result.Source,
	}

	var assignments []classification.TagAssignment
	var runModel string
	var runErr error

	switch opts.Provider {
	case "onnx":
		runModel = opts.ModelName
		assignments, runErr = s.onnxMgr.ClassifyItem(ctx, item, opts)
	case "llm":
		runModel = opts.LLMModel
		assignments, runErr = s.llmCls.ClassifyItem(ctx, item, opts)
	default:
		_ = s.clsRepo.SkipTask(task.ID, "unsupported provider: "+opts.Provider)
		return true
	}

	if runErr != nil {
		slog.Warn("Classification task failed", "task_id", task.ID, "module_id", task.ModuleID, "provider", opts.Provider, "error", runErr)
		_ = s.clsRepo.FailTask(task.ID, opts.Provider, runModel, runErr.Error())
		return true
	}

	if len(assignments) == 0 {
		errMsg := "no tags assigned (confidence threshold not met)"
		slog.Warn("Classification task yielded zero tags", "task_id", task.ID, "module_id", task.ModuleID, "provider", opts.Provider, "model", runModel)
		_ = s.clsRepo.FailTask(task.ID, opts.Provider, runModel, errMsg)
		return true
	}

	if err := s.clsRepo.ReplaceResultTags(task.ResultID, assignments, true); err != nil {
		_ = s.clsRepo.FailTask(task.ID, opts.Provider, runModel, err.Error())
		return true
	}

	_ = s.clsRepo.CompleteTask(task.ID, opts.Provider, runModel, len(assignments))
	return true
}

func (s *ClassificationService) resolveEffectiveOptions(moduleID string) (classification.EngineOptions, error) {
	sysMap, err := s.cfgRepo.GetMap()
	if err != nil {
		return classification.EngineOptions{}, err
	}

	conf, _ := strconv.Atoi(sysMap["classification_confidence"])
	if conf <= 0 {
		conf = 60
	}
	maxInputChars := classification.DefaultMaxInputChars
	if rawLimit, exists := sysMap["classification_max_input_chars"]; exists && strings.TrimSpace(rawLimit) != "" {
		if parsed, err := strconv.Atoi(strings.TrimSpace(rawLimit)); err == nil && parsed >= 0 {
			maxInputChars = parsed
		}
	}
	truncateStrategy := classification.NormalizeTruncateStrategy(sysMap["classification_truncate_strategy"])

	opts := classification.EngineOptions{
		Provider:         strings.ToLower(strings.TrimSpace(sysMap["classification_provider"])),
		ModelName:        strings.TrimSpace(sysMap["classification_onnx_model"]),
		Confidence:       conf,
		MaxInputChars:    maxInputChars,
		TruncateStrategy: truncateStrategy,
		LLMProtocol:      strings.ToLower(strings.TrimSpace(sysMap["classification_llm_protocol"])),
		LLMURL:           strings.TrimSpace(sysMap["classification_llm_url"]),
		LLMAPIKey:        strings.TrimSpace(sysMap["classification_llm_api_key"]),
		LLMModel:         strings.TrimSpace(sysMap["classification_llm_model"]),
		LLMPrompt:        strings.TrimSpace(sysMap["classification_llm_prompt"]),
	}
	if opts.Provider == "" {
		opts.Provider = "none"
	}
	if opts.LLMProtocol == "" {
		opts.LLMProtocol = "ollama"
	}
	if opts.LLMURL == "" {
		opts.LLMURL = "http://127.0.0.1:11434"
	}

	var profile *model.TagProfile
	if moduleID != "" {
		modCfg, err := s.clsRepo.GetModuleConfig(moduleID)
		if err == nil && modCfg != nil && modCfg.OverrideSystem {
			opts.Provider = strings.ToLower(strings.TrimSpace(modCfg.Provider))
			if modCfg.Confidence > 0 {
				opts.Confidence = modCfg.Confidence
			}
			if modCfg.MaxInputChars != nil && *modCfg.MaxInputChars >= 0 {
				opts.MaxInputChars = *modCfg.MaxInputChars
			}
			if strings.TrimSpace(modCfg.TruncateStrategy) != "" {
				opts.TruncateStrategy = classification.NormalizeTruncateStrategy(modCfg.TruncateStrategy)
			}
			if strings.TrimSpace(modCfg.ModelName) != "" {
				opts.ModelName = strings.TrimSpace(modCfg.ModelName)
			}
			if strings.TrimSpace(modCfg.LLMProtocol) != "" {
				opts.LLMProtocol = strings.ToLower(strings.TrimSpace(modCfg.LLMProtocol))
			}
			if strings.TrimSpace(modCfg.LLMURL) != "" {
				opts.LLMURL = strings.TrimSpace(modCfg.LLMURL)
			}
			if strings.TrimSpace(modCfg.LLMAPIKey) != "" {
				opts.LLMAPIKey = strings.TrimSpace(modCfg.LLMAPIKey)
			}
			if strings.TrimSpace(modCfg.LLMModel) != "" {
				opts.LLMModel = strings.TrimSpace(modCfg.LLMModel)
			}
			if strings.TrimSpace(modCfg.LLMPrompt) != "" {
				opts.LLMPrompt = strings.TrimSpace(modCfg.LLMPrompt)
			}
			if modCfg.ProfileID != nil && *modCfg.ProfileID > 0 {
				if p, pErr := s.clsRepo.GetProfile(*modCfg.ProfileID); pErr == nil {
					profile = p
				}
			}
		}
	}

	if profile == nil {
		if p, pErr := s.clsRepo.GetDefaultProfile(); pErr == nil {
			profile = p
		}
	}
	opts.Profile = repository.ProfileToSpec(profile)
	return opts, nil
}

func (s *ClassificationService) EnqueueSavedResults(records []model.ScraperResult) (int, error) {
	if len(records) == 0 {
		return 0, nil
	}
	pairs := make([]struct {
		ResultID uint
		ModuleID string
	}, 0, len(records))
	for _, r := range records {
		pairs = append(pairs, struct {
			ResultID uint
			ModuleID string
		}{ResultID: r.ID, ModuleID: r.ModuleID})
	}
	added, err := s.clsRepo.EnqueueResults(pairs)
	if err == nil && added > 0 {
		s.Wake()
	}
	return added, err
}

func (s *ClassificationService) EnqueueUntagged(limit int) (int, error) {
	items, err := s.clsRepo.ListUntaggedResults(limit)
	if err != nil {
		return 0, common.NewAPIError(http.StatusInternalServerError, common.ErrDatabaseError)
	}
	added, err := s.clsRepo.EnqueueResults(items)
	if err != nil {
		return 0, common.NewAPIError(http.StatusInternalServerError, common.ErrDatabaseError)
	}
	if added > 0 {
		s.Wake()
	}
	return added, nil
}

func (s *ClassificationService) EnqueueBatch(pairs []struct {
	ResultID uint
	ModuleID string
}) (int, error) {
	added, err := s.clsRepo.EnqueueResults(pairs)
	if err != nil {
		return 0, common.NewAPIError(http.StatusInternalServerError, common.ErrDatabaseError)
	}
	if added > 0 {
		s.Wake()
	}
	return added, nil
}

func (s *ClassificationService) GetOverview() (*ClassificationOverviewView, error) {
	sysMap, _ := s.cfgRepo.GetMap()
	provider := strings.TrimSpace(sysMap["classification_provider"])
	if provider == "" {
		provider = "none"
	}
	conf, _ := strconv.Atoi(sysMap["classification_confidence"])
	if conf <= 0 {
		conf = 60
	}
	localModels, _ := s.onnxMgr.ListLocalModels()
	stats, _ := s.clsRepo.GetQueueStats()
	state, curID := s.WorkerState()

	return &ClassificationOverviewView{
		SystemProvider:   provider,
		SystemConfidence: conf,
		ONNXEnv:          s.onnxMgr.EnvStatus(),
		ONNXPresetGroups: s.onnxMgr.Presets(),
		ONNXLocalModels:  localModels,
		LLMPresetModels:  classification.PresetLLMModels(),
		Queue: QueueStatusView{
			WorkerState:   state,
			Paused:        s.paused.Load(),
			CurrentTaskID: curID,
			Stats:         stats,
		},
		Jobs: s.ListJobs(),
	}, nil
}

func cloneJobCopy(j ClassificationJob) ClassificationJob {
	if j.FinishedAt != nil {
		t := *j.FinishedAt
		j.FinishedAt = &t
	}
	return j
}

func (s *ClassificationService) ListJobs() []ClassificationJob {
	s.jobsMu.RLock()
	defer s.jobsMu.RUnlock()
	if len(s.jobOrder) == 0 {
		return []ClassificationJob{}
	}
	out := make([]ClassificationJob, 0, len(s.jobOrder))
	for i := len(s.jobOrder) - 1; i >= 0; i-- {
		if entry, ok := s.jobs[s.jobOrder[i]]; ok {
			out = append(out, cloneJobCopy(entry.job))
		}
	}
	return out
}

func (s *ClassificationService) GetJob(id string) (*ClassificationJob, error) {
	s.jobsMu.RLock()
	defer s.jobsMu.RUnlock()
	entry, ok := s.jobs[id]
	if !ok {
		return nil, common.NewAPIError(http.StatusNotFound, common.ErrRecordNotFound)
	}
	cp := cloneJobCopy(entry.job)
	return &cp, nil
}

func (s *ClassificationService) CancelJob(id string) (*ClassificationJob, error) {
	s.jobsMu.Lock()
	entry, ok := s.jobs[id]
	if !ok {
		s.jobsMu.Unlock()
		return nil, common.NewAPIError(http.StatusNotFound, common.ErrRecordNotFound)
	}
	var cancelFn context.CancelFunc
	if entry.job.Status == ClassJobRunning {
		now := time.Now().UTC()
		entry.job.Status = ClassJobCancelled
		entry.job.Stage = "cancelled"
		entry.job.Indeterminate = false
		entry.job.UpdatedAt = now
		entry.job.FinishedAt = &now
		cancelFn = entry.cancel
	}
	cp := cloneJobCopy(entry.job)
	s.jobsMu.Unlock()

	if cancelFn != nil {
		cancelFn()
	}
	return &cp, nil
}

func (s *ClassificationService) DismissJob(id string) error {
	s.jobsMu.Lock()
	defer s.jobsMu.Unlock()
	entry, ok := s.jobs[id]
	if !ok {
		return nil
	}
	if entry.job.Status == ClassJobRunning {
		return common.NewAPIError(http.StatusConflict, common.ErrOperationFailed)
	}
	delete(s.jobs, id)
	nextOrder := make([]string, 0, len(s.jobOrder))
	for _, jid := range s.jobOrder {
		if jid != id {
			nextOrder = append(nextOrder, jid)
		}
	}
	s.jobOrder = nextOrder
	return nil
}

func (s *ClassificationService) launchAsyncJob(
	kind, target, endpointURL, initialStage, initialMsg string,
	run func(ctx context.Context, onProgress func(classification.ProgressUpdate)) (output string, err error),
) ClassificationJob {
	s.jobsMu.Lock()
	if s.jobs == nil {
		s.jobs = make(map[string]*classJobEntry)
	}
	for _, id := range s.jobOrder {
		if existing, ok := s.jobs[id]; ok {
			if existing.job.Kind == kind && existing.job.Target == target && existing.job.Status == ClassJobRunning {
				cp := cloneJobCopy(existing.job)
				s.jobsMu.Unlock()
				return cp
			}
		}
	}

	seq := s.jobSeq.Add(1)
	jobID := fmt.Sprintf("cls_%d_%d", time.Now().UTC().UnixMilli(), seq)
	baseCtx := s.bgCtx
	if baseCtx == nil {
		baseCtx = context.Background()
	}
	ctx, cancel := context.WithCancel(baseCtx)
	now := time.Now().UTC()

	entry := &classJobEntry{
		job: ClassificationJob{
			ID:            jobID,
			Kind:          kind,
			Target:        target,
			EndpointURL:   endpointURL,
			Status:        ClassJobRunning,
			Stage:         initialStage,
			Message:       initialMsg,
			CurrentItem:   target,
			Percent:       -1,
			Indeterminate: true,
			CreatedAt:     now,
			UpdatedAt:     now,
		},
		cancel: cancel,
	}
	s.jobs[jobID] = entry
	s.jobOrder = append(s.jobOrder, jobID)
	s.pruneJobsLocked(now)
	cp := cloneJobCopy(entry.job)
	s.jobsWG.Add(1)
	s.jobsMu.Unlock()

	go func() {
		defer s.jobsWG.Done()
		defer cancel()

		onProgress := func(upd classification.ProgressUpdate) {
			s.jobsMu.Lock()
			defer s.jobsMu.Unlock()
			e, ok := s.jobs[jobID]
			if !ok || e.job.Status != ClassJobRunning {
				return
			}
			ts := time.Now().UTC()
			if upd.Stage != "" {
				e.job.Stage = upd.Stage
			}
			if upd.Message != "" {
				e.job.Message = upd.Message
			}
			if upd.CurrentItem != "" {
				e.job.CurrentItem = upd.CurrentItem
			}
			if upd.CompletedBytes > 0 || upd.TotalBytes > 0 {
				e.job.CompletedBytes = upd.CompletedBytes
				e.job.TotalBytes = upd.TotalBytes
			}
			e.job.Percent = upd.Percent
			e.job.Indeterminate = upd.Indeterminate
			e.job.UpdatedAt = ts
		}

		out, runErr := run(ctx, onProgress)

		s.jobsMu.Lock()
		defer s.jobsMu.Unlock()
		e, ok := s.jobs[jobID]
		if !ok {
			return
		}
		if e.job.Status == ClassJobCancelled {
			return
		}
		doneAt := time.Now().UTC()
		e.job.UpdatedAt = doneAt
		e.job.FinishedAt = &doneAt
		e.job.Output = strings.TrimSpace(out)
		e.job.Indeterminate = false
		if runErr != nil {
			if errors.Is(runErr, context.Canceled) && ctx.Err() != nil {
				e.job.Status = ClassJobCancelled
				e.job.Stage = "cancelled"
			} else {
				e.job.Status = ClassJobFailed
				e.job.Stage = "failed"
				e.job.Error = runErr.Error()
			}
			return
		}
		e.job.Status = ClassJobCompleted
		e.job.Stage = "ready"
		e.job.Percent = 100
		if e.job.TotalBytes > 0 && e.job.CompletedBytes < e.job.TotalBytes {
			e.job.CompletedBytes = e.job.TotalBytes
		}
	}()

	return cp
}

func (s *ClassificationService) pruneJobsLocked(now time.Time) {
	if len(s.jobOrder) <= 20 {
		return
	}
	kept := make([]string, 0, len(s.jobOrder))
	finishedToDrop := len(s.jobOrder) - 20
	for _, id := range s.jobOrder {
		entry, ok := s.jobs[id]
		if !ok {
			continue
		}
		if entry.job.Status != ClassJobRunning && (finishedToDrop > 0 || (entry.job.FinishedAt != nil && now.Sub(*entry.job.FinishedAt) > 2*time.Hour)) {
			delete(s.jobs, id)
			if finishedToDrop > 0 {
				finishedToDrop--
			}
			continue
		}
		kept = append(kept, id)
	}
	s.jobOrder = kept
}

func (s *ClassificationService) StartPrepareONNXEnvJob() ClassificationJob {
	return s.launchAsyncJob(
		ClassJobONNXEnv,
		"onnx_runtime",
		"",
		"preparing_env",
		"initializing python runtime",
		func(ctx context.Context, onProgress func(classification.ProgressUpdate)) (string, error) {
			out, err := s.onnxMgr.PrepareEnvWithProgress(ctx, onProgress)
			if err == nil && s.bus != nil {
				s.bus.Emit(model.EventLevelNormal, model.EventCategorySystem, "event.classification.onnx_env_ready", nil, nil, "classification", "onnx")
			}
			return out, err
		},
	)
}

func (s *ClassificationService) PrepareONNXEnv(ctx context.Context) (classification.ONNXEnvStatus, string, error) {
	out, err := s.onnxMgr.PrepareEnvWithProgress(ctx, nil)
	if err != nil {
		return s.onnxMgr.EnvStatus(), out, common.NewAPIError(http.StatusInternalServerError, common.ErrONNXEnvFailed, map[string]any{"detail": err.Error()})
	}
	if s.bus != nil {
		s.bus.Emit(model.EventLevelNormal, model.EventCategorySystem, "event.classification.onnx_env_ready", nil, nil, "classification", "onnx")
	}
	return s.onnxMgr.EnvStatus(), out, nil
}

func (s *ClassificationService) StartDownloadONNXModelJob(repoID string) (ClassificationJob, error) {
	clean := strings.TrimSpace(repoID)
	if clean == "" || !strings.Contains(clean, "/") || strings.Contains(clean, "..") {
		return ClassificationJob{}, common.NewAPIError(http.StatusBadRequest, common.ErrValidationFailed)
	}
	job := s.launchAsyncJob(
		ClassJobONNXDownload,
		clean,
		"",
		"resolving",
		"queued",
		func(ctx context.Context, onProgress func(classification.ProgressUpdate)) (string, error) {
			if err := s.onnxMgr.DownloadModelWithProgress(ctx, clean, onProgress); err != nil {
				return "", err
			}
			if s.bus != nil {
				s.bus.Emit(model.EventLevelNormal, model.EventCategorySystem, "event.classification.model_downloaded", map[string]any{"model": clean}, nil, "classification", "onnx")
			}
			return "", nil
		},
	)
	return job, nil
}

func (s *ClassificationService) DownloadONNXModel(ctx context.Context, repoID string) error {
	if err := s.onnxMgr.DownloadModelWithProgress(ctx, repoID, nil); err != nil {
		return common.NewAPIError(http.StatusBadGateway, common.ErrModelDownloadFailed, map[string]any{"detail": err.Error()})
	}
	if s.bus != nil {
		s.bus.Emit(model.EventLevelNormal, model.EventCategorySystem, "event.classification.model_downloaded", map[string]any{"model": repoID}, nil, "classification", "onnx")
	}
	return nil
}

func (s *ClassificationService) DeleteONNXModel(name string) error {
	if err := s.onnxMgr.DeleteModel(name); err != nil {
		return common.NewAPIError(http.StatusBadRequest, common.ErrValidationFailed)
	}
	return nil
}

func (s *ClassificationService) ListOllamaModels(ctx context.Context, urlOverride string) ([]classification.OllamaModelInfo, error) {
	baseURL := strings.TrimSpace(urlOverride)
	if baseURL == "" {
		sysMap, _ := s.cfgRepo.GetMap()
		baseURL = strings.TrimSpace(sysMap["classification_llm_url"])
	}
	if baseURL == "" {
		baseURL = "http://127.0.0.1:11434"
	}
	list, err := s.llmCls.ListOllamaModels(ctx, baseURL)
	if err != nil {
		return nil, common.NewAPIError(http.StatusBadGateway, common.ErrOllamaRequestFailed, map[string]any{"detail": err.Error()})
	}
	return list, nil
}

func (s *ClassificationService) StartPullOllamaModelJob(urlOverride, modelName string) (ClassificationJob, error) {
	cleanModel := strings.TrimSpace(modelName)
	if cleanModel == "" {
		return ClassificationJob{}, common.NewAPIError(http.StatusBadRequest, common.ErrValidationFailed)
	}
	baseURL := strings.TrimSpace(urlOverride)
	if baseURL == "" {
		sysMap, _ := s.cfgRepo.GetMap()
		baseURL = strings.TrimSpace(sysMap["classification_llm_url"])
	}
	if baseURL == "" {
		baseURL = "http://127.0.0.1:11434"
	}
	job := s.launchAsyncJob(
		ClassJobOllamaPull,
		cleanModel,
		baseURL,
		"resolving",
		"pulling manifest",
		func(ctx context.Context, onProgress func(classification.ProgressUpdate)) (string, error) {
			if err := s.llmCls.PullOllamaModel(ctx, baseURL, cleanModel, onProgress); err != nil {
				return "", err
			}
			if s.bus != nil {
				s.bus.Emit(model.EventLevelNormal, model.EventCategorySystem, "event.classification.model_downloaded", map[string]any{"model": cleanModel}, nil, "classification", "ollama")
			}
			return "", nil
		},
	)
	return job, nil
}

func (s *ClassificationService) PullOllamaModel(ctx context.Context, urlOverride, modelName string) error {
	baseURL := strings.TrimSpace(urlOverride)
	if baseURL == "" {
		sysMap, _ := s.cfgRepo.GetMap()
		baseURL = strings.TrimSpace(sysMap["classification_llm_url"])
	}
	if baseURL == "" {
		baseURL = "http://127.0.0.1:11434"
	}
	if err := s.llmCls.PullOllamaModel(ctx, baseURL, modelName, nil); err != nil {
		return common.NewAPIError(http.StatusBadGateway, common.ErrOllamaRequestFailed, map[string]any{"detail": err.Error()})
	}
	if s.bus != nil {
		s.bus.Emit(model.EventLevelNormal, model.EventCategorySystem, "event.classification.model_downloaded", map[string]any{"model": modelName}, nil, "classification", "ollama")
	}
	return nil
}

func (s *ClassificationService) ListProfiles() ([]model.TagProfile, error) {
	list, err := s.clsRepo.ListProfiles()
	if err != nil {
		return nil, common.NewAPIError(http.StatusInternalServerError, common.ErrDatabaseError)
	}
	return list, nil
}

func (s *ClassificationService) CreateProfile(input repository.ProfileSaveInput) (*model.TagProfile, error) {
	p, err := s.clsRepo.CreateProfile(input)
	if err != nil {
		return nil, common.SanitizeDBError(err)
	}
	return p, nil
}

func (s *ClassificationService) UpdateProfile(id uint, input repository.ProfileSaveInput) (*model.TagProfile, error) {
	p, err := s.clsRepo.UpdateProfile(id, input)
	if err != nil {
		return nil, common.SanitizeDBError(err)
	}
	return p, nil
}

func (s *ClassificationService) DeleteProfile(id uint) error {
	if err := s.clsRepo.DeleteProfile(id); err != nil {
		return common.NewAPIError(http.StatusBadRequest, common.ErrOperationFailed, map[string]any{"detail": err.Error()})
	}
	return nil
}

func (s *ClassificationService) LoadDefaultProfileTemplate(id uint, locale string) (*model.TagProfile, error) {
	existing, err := s.clsRepo.GetProfile(id)
	if err != nil {
		return nil, common.NewAPIError(http.StatusNotFound, common.ErrClassificationProfileNotFound)
	}
	tpl := classification.DefaultTemplate(locale)
	groups := make([]repository.ProfileGroupInput, 0, len(tpl.Groups))
	for _, g := range tpl.Groups {
		tags := make([]repository.ProfileTagInput, 0, len(g.Tags))
		for _, t := range g.Tags {
			tags = append(tags, repository.ProfileTagInput{
				Name:        t.Name,
				Description: t.Description,
				Priority:    t.Priority,
				Keywords:    t.Keywords,
			})
		}
		groups = append(groups, repository.ProfileGroupInput{
			Name:          g.Name,
			Description:   g.Description,
			SelectionMode: g.SelectionMode,
			MaxResults:    g.MaxResults,
			Tags:          tags,
		})
	}
	updated, err := s.clsRepo.UpdateProfile(id, repository.ProfileSaveInput{
		Name:        existing.Name,
		Description: existing.Description,
		IsDefault:   existing.IsDefault,
		Groups:      groups,
	})
	if err != nil {
		return nil, common.SanitizeDBError(err)
	}
	return updated, nil
}

func (s *ClassificationService) GetModuleConfig(moduleID string) (*ModuleClassificationConfigView, error) {
	if _, err := s.modRepo.GetByID(moduleID); err != nil {
		return nil, common.NewAPIError(http.StatusNotFound, common.ErrModuleNotFound)
	}
	cfg, err := s.clsRepo.GetModuleConfig(moduleID)
	if err != nil {
		return nil, common.NewAPIError(http.StatusInternalServerError, common.ErrDatabaseError)
	}
	keyView := ""
	if cfg.LLMAPIKey != "" {
		keyView = "********"
	}
	return &ModuleClassificationConfigView{
		ModuleID:         cfg.ModuleID,
		OverrideSystem:   cfg.OverrideSystem,
		Provider:         cfg.Provider,
		ModelName:        cfg.ModelName,
		Confidence:       cfg.Confidence,
		MaxInputChars:    cfg.MaxInputChars,
		TruncateStrategy: cfg.TruncateStrategy,
		ProfileID:        cfg.ProfileID,
		LLMProtocol:      cfg.LLMProtocol,
		LLMURL:           cfg.LLMURL,
		LLMAPIKey:        keyView,
		LLMModel:         cfg.LLMModel,
		LLMPrompt:        cfg.LLMPrompt,
	}, nil
}

func (s *ClassificationService) SaveModuleConfig(moduleID string, in ModuleClassificationConfigView) error {
	if _, err := s.modRepo.GetByID(moduleID); err != nil {
		return common.NewAPIError(http.StatusNotFound, common.ErrModuleNotFound)
	}
	existing, _ := s.clsRepo.GetModuleConfig(moduleID)
	apiKey := strings.TrimSpace(in.LLMAPIKey)
	if apiKey == "********" && existing != nil {
		apiKey = existing.LLMAPIKey
	}
	conf := in.Confidence
	if conf <= 0 {
		conf = 60
	}
	if conf > 100 {
		conf = 100
	}
	prov := strings.ToLower(strings.TrimSpace(in.Provider))
	if prov != "none" && prov != "onnx" && prov != "llm" {
		prov = "none"
	}
	proto := strings.ToLower(strings.TrimSpace(in.LLMProtocol))
	if proto != "ollama" && proto != "openai" {
		proto = "ollama"
	}
	var maxChars *int
	if in.MaxInputChars != nil && *in.MaxInputChars >= 0 {
		v := *in.MaxInputChars
		maxChars = &v
	}
	truncStrategy := strings.ToLower(strings.TrimSpace(in.TruncateStrategy))
	if truncStrategy != classification.TruncateStrategyTitleHeadTail &&
		truncStrategy != classification.TruncateStrategyHead &&
		truncStrategy != classification.TruncateStrategyTail {
		truncStrategy = ""
	}
	rec := &model.ModuleClassificationConfig{
		ModuleID:         moduleID,
		OverrideSystem:   in.OverrideSystem,
		Provider:         prov,
		ModelName:        strings.TrimSpace(in.ModelName),
		Confidence:       conf,
		MaxInputChars:    maxChars,
		TruncateStrategy: truncStrategy,
		ProfileID:        in.ProfileID,
		LLMProtocol:      proto,
		LLMURL:           strings.TrimSpace(in.LLMURL),
		LLMAPIKey:        apiKey,
		LLMModel:         strings.TrimSpace(in.LLMModel),
		LLMPrompt:        strings.TrimSpace(in.LLMPrompt),
	}
	if err := s.clsRepo.UpsertModuleConfig(rec); err != nil {
		return common.NewAPIError(http.StatusInternalServerError, common.ErrDatabaseError)
	}
	return nil
}

func (s *ClassificationService) GetQueueStatus() (QueueStatusView, error) {
	stats, err := s.clsRepo.GetQueueStats()
	if err != nil {
		return QueueStatusView{}, common.NewAPIError(http.StatusInternalServerError, common.ErrDatabaseError)
	}
	state, curID := s.WorkerState()
	return QueueStatusView{
		WorkerState:   state,
		Paused:        s.paused.Load(),
		CurrentTaskID: curID,
		Stats:         stats,
	}, nil
}

func (s *ClassificationService) ListQueueTasks(status string, page, pageSize int) ([]repository.ClassificationTaskView, int64, error) {
	items, total, err := s.clsRepo.ListTasks(status, page, pageSize)
	if err != nil {
		return nil, 0, common.NewAPIError(http.StatusInternalServerError, common.ErrDatabaseError)
	}
	return items, total, nil
}

func (s *ClassificationService) GetQueueTaskDetail(taskID uint) (*repository.ClassificationTaskDetailView, error) {
	detail, err := s.clsRepo.GetTaskDetail(taskID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, common.NewAPIError(http.StatusNotFound, common.ErrRecordNotFound)
		}
		return nil, common.NewAPIError(http.StatusInternalServerError, common.ErrDatabaseError)
	}
	return detail, nil
}

func (s *ClassificationService) CancelPendingTasks() (int64, error) {
	cnt, err := s.clsRepo.CancelPending()
	if err != nil {
		return 0, common.NewAPIError(http.StatusInternalServerError, common.ErrDatabaseError)
	}
	return cnt, nil
}

func (s *ClassificationService) CancelTask(taskID uint) error {
	if err := s.clsRepo.CancelTask(taskID); err != nil {
		return common.NewAPIError(http.StatusInternalServerError, common.ErrDatabaseError)
	}
	return nil
}

func (s *ClassificationService) RetryTask(taskID uint) error {
	if err := s.clsRepo.RetryTask(taskID); err != nil {
		return common.NewAPIError(http.StatusInternalServerError, common.ErrDatabaseError)
	}
	s.Wake()
	return nil
}

func (s *ClassificationService) RetryFailedTasks() (int64, error) {
	cnt, err := s.clsRepo.RetryFailed()
	if err != nil {
		return 0, common.NewAPIError(http.StatusInternalServerError, common.ErrDatabaseError)
	}
	if cnt > 0 {
		s.Wake()
	}
	return cnt, nil
}

func (s *ClassificationService) ClearFinishedTasks() (int64, error) {
	cnt, err := s.clsRepo.ClearFinished()
	if err != nil {
		return 0, common.NewAPIError(http.StatusInternalServerError, common.ErrDatabaseError)
	}
	return cnt, nil
}

func (s *ClassificationService) ListAllTags() ([]model.Tag, error) {
	tags, err := s.clsRepo.ListAllTags()
	if err != nil {
		return nil, common.NewAPIError(http.StatusInternalServerError, common.ErrDatabaseError)
	}
	return tags, nil
}
