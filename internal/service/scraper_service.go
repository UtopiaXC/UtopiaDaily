package service

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/UtopiaXC/UtopiaDaily/internal/classification"
	"github.com/UtopiaXC/UtopiaDaily/internal/common"
	"github.com/UtopiaXC/UtopiaDaily/internal/event"
	"github.com/UtopiaXC/UtopiaDaily/internal/model"
	"github.com/UtopiaXC/UtopiaDaily/internal/module"
	"github.com/UtopiaXC/UtopiaDaily/internal/module/external"
	"github.com/UtopiaXC/UtopiaDaily/internal/module/registry"
	"github.com/UtopiaXC/UtopiaDaily/internal/repository"
)

type ModuleFieldView struct {
	Key            string   `json:"key"`
	Type           string   `json:"type"`
	Value          any      `json:"value"`
	Description    string   `json:"description"`
	Hint           string   `json:"hint,omitempty"`
	Options        []string `json:"options,omitempty"`
	Secret         bool     `json:"secret,omitempty"`
	OverrideSystem bool     `json:"override_system,omitempty"`
	SortOrder      int      `json:"sort_order"`
}

type ModuleTaskView struct {
	ID          uint       `json:"id"`
	TaskKey     string     `json:"task_key"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	CronExpr    string     `json:"cron_expr"`
	Enabled     bool       `json:"enabled"`
	LastRunAt   *time.Time `json:"last_run_at,omitempty"`
	LastStatus  string     `json:"last_status"`
	LastError   string     `json:"last_error,omitempty"`
}

type ModuleSummaryView struct {
	ID              string    `json:"module_id"`
	Kind            string    `json:"kind"`
	Source          string    `json:"source"`
	Name            string    `json:"name"`
	Description     string    `json:"description"`
	Version         string    `json:"version"`
	Author          string    `json:"author"`
	Enabled         bool      `json:"is_enable"`
	EnvReady        bool      `json:"env_ready"`
	HasRequirements bool      `json:"has_requirements"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type ModuleDetailView struct {
	ID              string            `json:"module_id"`
	Kind            string            `json:"kind"`
	Source          string            `json:"source"`
	Name            string            `json:"name"`
	Description     string            `json:"description"`
	Version         string            `json:"version"`
	Author          string            `json:"author"`
	Enabled         bool              `json:"is_enable"`
	EnvReady        bool              `json:"env_ready"`
	HasRequirements bool              `json:"has_requirements"`
	Config          []ModuleFieldView `json:"config"`
	Tasks           []ModuleTaskView  `json:"tasks"`
}

type ScraperService struct {
	reg         *registry.Registry
	modRepo     *repository.ModuleRepository
	scraperRepo *repository.ScraperRepository
	cfgRepo     *repository.ConfigRepository
	classifier  classification.Classifier
	bus         *event.Bus
	onReload    func()
}

func NewScraperService(
	reg *registry.Registry,
	modRepo *repository.ModuleRepository,
	scraperRepo *repository.ScraperRepository,
	cfgRepo *repository.ConfigRepository,
	classifier classification.Classifier,
	bus *event.Bus,
) *ScraperService {
	return &ScraperService{
		reg:         reg,
		modRepo:     modRepo,
		scraperRepo: scraperRepo,
		cfgRepo:     cfgRepo,
		classifier:  classifier,
		bus:         bus,
	}
}

func (s *ScraperService) SetOnScheduleReload(fn func()) {
	s.onReload = fn
}

func (s *ScraperService) SyncAllModules() error {
	if err := s.reg.ReloadAll(); err != nil {
		return err
	}
	for _, m := range s.reg.AllModules() {
		meta := m.Meta()
		envReady := true
		if meta.Source == module.SourceExternal && meta.EntryDir != "" {
			if external.HasRequirements(meta.EntryDir) {
				envReady = external.IsVenvReady(meta.EntryDir)
			}
		}
		if err := s.modRepo.SyncModule(m, envReady, "", ""); err != nil {
			return err
		}
	}
	return nil
}

func (s *ScraperService) ReloadAll() error {
	return reloadAllModulesCommon(s.reg, s.modRepo, s.bus, s.onReload)
}

func (s *ScraperService) ListModules() ([]ModuleSummaryView, error) {
	list, err := s.modRepo.ListByKind(module.KindScraper)
	if err != nil {
		return nil, common.NewAPIError(http.StatusInternalServerError, common.ErrDatabaseError)
	}
	return buildModuleSummaryList(s.reg, list), nil
}

func (s *ScraperService) GetModuleDetail(id string) (*ModuleDetailView, error) {
	modRec, err := s.modRepo.GetByID(id)
	if err != nil || modRec.Kind != module.KindScraper {
		return nil, common.NewAPIError(http.StatusNotFound, common.ErrModuleNotFound)
	}
	contract, ok := s.reg.GetModule(id)
	if !ok {
		return nil, common.NewAPIError(http.StatusNotFound, common.ErrModuleNotFound)
	}
	rows, _ := s.modRepo.GetConfigRows(id)
	jobs, _ := s.modRepo.ListJobsByModule(id)
	return buildModuleDetailView(modRec, contract, rows, jobs), nil
}

func (s *ScraperService) PrepareEnv(ctx context.Context, id string) (string, error) {
	return prepareEnvCommon(ctx, s.reg, s.modRepo, s.bus, module.KindScraper, id)
}

func (s *ScraperService) EnableModule(ctx context.Context, id string) error {
	return enableModuleCommon(ctx, s.reg, s.modRepo, s.bus, s.onReload, module.KindScraper, id)
}

func (s *ScraperService) DisableModule(id string) error {
	return disableModuleCommon(s.modRepo, s.bus, s.onReload, module.KindScraper, id)
}

func (s *ScraperService) SaveConfig(id string, payload map[string]any) error {
	contract, ok := s.reg.GetScraper(id)
	if !ok {
		return common.NewAPIError(http.StatusNotFound, common.ErrModuleNotFound)
	}
	if err := s.modRepo.SaveConfigValues(id, contract.ConfigSchema(), payload); err != nil {
		return common.NewAPIError(http.StatusInternalServerError, common.ErrDatabaseError)
	}
	return nil
}

func (s *ScraperService) TestConfig(ctx context.Context, id string, overrides map[string]any) (module.TestResult, error) {
	contract, ok := s.reg.GetScraper(id)
	if !ok {
		return module.TestResult{}, common.NewAPIError(http.StatusNotFound, common.ErrModuleNotFound)
	}
	rows, _ := s.modRepo.GetConfigRows(id)
	sysMap, _ := s.cfgRepo.GetMap()
	acc := repository.NewMemoryConfigAccessor(contract.ConfigSchema(), overrides, rows, sysMap)
	return contract.TestConfig(ctx, acc), nil
}

func (s *ScraperService) RunTask(ctx context.Context, id, taskKey string) (int, error) {
	contract, ok := s.reg.GetScraper(id)
	if !ok {
		return 0, common.NewAPIError(http.StatusNotFound, common.ErrModuleNotFound)
	}
	modRec, err := s.modRepo.GetByID(id)
	if err != nil {
		return 0, common.NewAPIError(http.StatusNotFound, common.ErrModuleNotFound)
	}
	if meta := contract.Meta(); meta.Source == module.SourceExternal && !modRec.EnvReady && external.HasRequirements(meta.EntryDir) {
		if _, err := s.PrepareEnv(ctx, id); err != nil {
			_ = s.modRepo.RecordScheduleRun(id, taskKey, "failed", err.Error())
			return 0, err
		}
	}

	rows, _ := s.modRepo.GetConfigRows(id)
	sysMap, _ := s.cfgRepo.GetMap()
	acc := repository.NewDBConfigAccessor(contract.ConfigSchema(), rows, sysMap)

	items, scrapeErr := contract.Scrape(ctx, taskKey, acc)
	if scrapeErr != nil {
		_ = s.modRepo.RecordScheduleRun(id, taskKey, "failed", scrapeErr.Error())
		if s.bus != nil {
			s.bus.Emit(model.EventLevelCritical, model.EventCategoryTask, "event.scraper.failed", map[string]any{
				"module": id,
				"task":   taskKey,
			}, map[string]any{"error": scrapeErr.Error()}, "scraper", id)
		}
		return 0, common.NewAPIError(http.StatusBadGateway, common.ErrModuleExecutionFailed, map[string]any{"detail": scrapeErr.Error()})
	}

	inserted, savedRecords, saveErr := s.scraperRepo.SaveScrapedItems(id, taskKey, items)
	if saveErr != nil {
		_ = s.modRepo.RecordScheduleRun(id, taskKey, "failed", saveErr.Error())
		return 0, common.NewAPIError(http.StatusInternalServerError, common.ErrDatabaseError)
	}

	if inserted > 0 && s.classifier != nil {
		inputs := make([]classification.InputItem, 0, len(savedRecords))
		idByFP := make(map[string]uint, len(savedRecords))
		for _, rec := range savedRecords {
			inputs = append(inputs, classification.InputItem{
				Fingerprint: rec.Fingerprint,
				Title:       rec.Title,
				Summary:     rec.Summary,
				Content:     rec.Content,
				SourceName:  rec.Source,
			})
			idByFP[rec.Fingerprint] = rec.ID
		}
		if results, clsErr := s.classifier.ClassifyBatch(ctx, inputs); clsErr == nil {
			for _, cr := range results {
				if len(cr.Tags) > 0 {
					if rid, exists := idByFP[cr.Fingerprint]; exists {
						_ = s.scraperRepo.AttachTags(rid, cr.Tags)
					}
				}
			}
		}
	}

	_ = s.modRepo.RecordScheduleRun(id, taskKey, "success", "")
	if s.bus != nil {
		s.bus.Emit(model.EventLevelNormal, model.EventCategoryTask, "event.scraper.completed", map[string]any{
			"module":   id,
			"task":     taskKey,
			"inserted": inserted,
		}, map[string]any{"total_fetched": len(items), "inserted": inserted}, "scraper", id)
	}
	return inserted, nil
}

func buildModuleDetailView(modRec *model.Module, contract module.Module, rows []model.ModuleConfig, jobs []model.ScheduleJob) *ModuleDetailView {
	rowMap := make(map[string]model.ModuleConfig, len(rows))
	for _, r := range rows {
		rowMap[r.ConfigKey] = r
	}

	schema := contract.ConfigSchema()
	fields := make([]ModuleFieldView, 0, len(schema))
	for _, spec := range schema {
		var val any = spec.DefaultValue
		if spec.Type == "array" {
			val = spec.DefaultItems
		} else if spec.Type == "bool" || spec.Type == "boolean" {
			val = spec.DefaultValue == "true"
		}

		if row, ok := rowMap[spec.Key]; ok {
			switch spec.Type {
			case "array":
				arr := make([]string, 0, len(row.Items))
				for _, it := range row.Items {
					arr = append(arr, it.ItemValue)
				}
				val = arr
			case "bool", "boolean":
				val = row.ScalarValue == "true" || row.ScalarValue == "1"
			case "int":
				val = row.ScalarValue
			default:
				if spec.Secret && row.ScalarValue != "" {
					val = "********"
				} else {
					val = row.ScalarValue
				}
			}
		} else if spec.Secret && spec.DefaultValue != "" {
			val = "********"
		}

		fields = append(fields, ModuleFieldView{
			Key:            spec.Key,
			Type:           spec.Type,
			Value:          val,
			Description:    fmt.Sprintf("module.%s.config.%s.desc", modRec.ID, spec.Key),
			Hint:           fmt.Sprintf("module.%s.config.%s.hint", modRec.ID, spec.Key),
			Options:        spec.Options,
			Secret:         spec.Secret,
			OverrideSystem: spec.OverrideSystem,
			SortOrder:      spec.SortOrder,
		})
	}

	tasks := make([]ModuleTaskView, 0, len(jobs))
	for _, j := range jobs {
		tasks = append(tasks, ModuleTaskView{
			ID:          j.ID,
			TaskKey:     j.TaskKey,
			Name:        j.NameKey,
			Description: j.DescriptionKey,
			CronExpr:    j.CronExpr,
			Enabled:     j.Enabled,
			LastRunAt:   j.LastRunAt,
			LastStatus:  j.LastStatus,
			LastError:   j.LastError,
		})
	}

	hasReq := false
	if meta := contract.Meta(); meta.Source == module.SourceExternal && meta.EntryDir != "" {
		hasReq = external.HasRequirements(meta.EntryDir)
	}

	return &ModuleDetailView{
		ID:              modRec.ID,
		Kind:            modRec.Kind,
		Source:          modRec.Source,
		Name:            modRec.NameKey,
		Description:     modRec.DescriptionKey,
		Version:         modRec.Version,
		Author:          modRec.Author,
		Enabled:         modRec.Enabled,
		EnvReady:        modRec.EnvReady,
		HasRequirements: hasReq,
		Config:          fields,
		Tasks:           tasks,
	}
}

func buildModuleSummaryList(reg *registry.Registry, list []model.Module) []ModuleSummaryView {
	out := make([]ModuleSummaryView, 0, len(list))
	for _, m := range list {
		hasReq := false
		if contract, ok := reg.GetModule(m.ID); ok {
			meta := contract.Meta()
			if meta.Source == module.SourceExternal && meta.EntryDir != "" {
				hasReq = external.HasRequirements(meta.EntryDir)
			}
		}
		out = append(out, ModuleSummaryView{
			ID:              m.ID,
			Kind:            m.Kind,
			Source:          m.Source,
			Name:            m.NameKey,
			Description:     m.DescriptionKey,
			Version:         m.Version,
			Author:          m.Author,
			Enabled:         m.Enabled,
			EnvReady:        m.EnvReady,
			HasRequirements: hasReq,
			CreatedAt:       m.CreatedAt,
			UpdatedAt:       m.UpdatedAt,
		})
	}
	return out
}

func reloadAllModulesCommon(
	reg *registry.Registry,
	modRepo *repository.ModuleRepository,
	bus *event.Bus,
	onReload func(),
) error {
	if err := reg.ReloadAll(); err != nil {
		return common.NewAPIError(http.StatusInternalServerError, common.ErrOperationFailed)
	}
	for _, m := range reg.AllModules() {
		meta := m.Meta()
		envReady := true
		if meta.Source == module.SourceExternal && meta.EntryDir != "" {
			if external.HasRequirements(meta.EntryDir) {
				envReady = external.IsVenvReady(meta.EntryDir)
			}
		}
		if err := modRepo.SyncModule(m, envReady, "", ""); err != nil {
			return common.NewAPIError(http.StatusInternalServerError, common.ErrDatabaseError)
		}
	}
	if onReload != nil {
		onReload()
	}
	if bus != nil {
		bus.Emit(model.EventLevelNormal, model.EventCategoryModule, "event.module.reloaded", nil, nil, "module", "")
	}
	return nil
}

func prepareEnvCommon(
	ctx context.Context,
	reg *registry.Registry,
	modRepo *repository.ModuleRepository,
	bus *event.Bus,
	expectedKind, id string,
) (string, error) {
	contract, ok := reg.GetModule(id)
	if !ok {
		return "", common.NewAPIError(http.StatusNotFound, common.ErrModuleNotFound)
	}
	meta := contract.Meta()
	if meta.Kind != expectedKind {
		return "", common.NewAPIError(http.StatusNotFound, common.ErrModuleNotFound)
	}
	if meta.Source != module.SourceExternal || meta.EntryDir == "" {
		return "", nil
	}
	out, err := external.SetupVenv(ctx, meta.EntryDir)
	if err != nil {
		_ = modRepo.SetEnvReady(id, false)
		if bus != nil {
			bus.Emit(model.EventLevelCritical, model.EventCategoryModule, "event.module.venv_failed", map[string]any{
				"module": id,
			}, map[string]any{"detail": err.Error(), "output": out}, "module", id)
		}
		return out, common.NewAPIError(http.StatusInternalServerError, common.ErrPluginDependencyFailed, map[string]any{"detail": err.Error()})
	}
	_ = modRepo.SetEnvReady(id, true)
	if bus != nil {
		bus.Emit(model.EventLevelNormal, model.EventCategoryModule, "event.module.venv_ready", map[string]any{
			"module": id,
		}, nil, "module", id)
	}
	return out, nil
}

func enableModuleCommon(
	ctx context.Context,
	reg *registry.Registry,
	modRepo *repository.ModuleRepository,
	bus *event.Bus,
	onReload func(),
	expectedKind, id string,
) error {
	modRec, err := modRepo.GetByID(id)
	if err != nil || modRec.Kind != expectedKind {
		return common.NewAPIError(http.StatusNotFound, common.ErrModuleNotFound)
	}
	contract, ok := reg.GetModule(id)
	if !ok {
		return common.NewAPIError(http.StatusNotFound, common.ErrModuleNotFound)
	}
	meta := contract.Meta()
	if meta.Source == module.SourceExternal && meta.EntryDir != "" {
		if external.HasRequirements(meta.EntryDir) && !external.IsVenvReady(meta.EntryDir) {
			if _, err := external.SetupVenv(ctx, meta.EntryDir); err != nil {
				_ = modRepo.SetEnvReady(id, false)
				if bus != nil {
					bus.Emit(model.EventLevelCritical, model.EventCategoryModule, "event.module.venv_failed", map[string]any{
						"module": id,
					}, map[string]any{"detail": err.Error()}, "module", id)
				}
				return common.NewAPIError(http.StatusInternalServerError, common.ErrPluginDependencyFailed, map[string]any{"detail": err.Error()})
			}
			_ = modRepo.SetEnvReady(id, true)
			if bus != nil {
				bus.Emit(model.EventLevelNormal, model.EventCategoryModule, "event.module.venv_ready", map[string]any{
					"module": id,
				}, nil, "module", id)
			}
		}
	}

	if err := modRepo.SetEnabled(id, true); err != nil {
		return common.NewAPIError(http.StatusInternalServerError, common.ErrDatabaseError)
	}
	if onReload != nil {
		onReload()
	}
	if bus != nil {
		bus.Emit(model.EventLevelNormal, model.EventCategoryModule, "event.module.enabled", map[string]any{
			"module": id,
		}, nil, "module", id)
	}
	return nil
}

func disableModuleCommon(
	modRepo *repository.ModuleRepository,
	bus *event.Bus,
	onReload func(),
	expectedKind, id string,
) error {
	modRec, err := modRepo.GetByID(id)
	if err != nil || modRec.Kind != expectedKind {
		return common.NewAPIError(http.StatusNotFound, common.ErrModuleNotFound)
	}
	if err := modRepo.SetEnabled(id, false); err != nil {
		return common.NewAPIError(http.StatusInternalServerError, common.ErrDatabaseError)
	}
	if onReload != nil {
		onReload()
	}
	if bus != nil {
		bus.Emit(model.EventLevelNormal, model.EventCategoryModule, "event.module.disabled", map[string]any{
			"module": id,
		}, nil, "module", id)
	}
	return nil
}
