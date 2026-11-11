package service

import (
	"context"
	"net/http"
	"time"

	"github.com/UtopiaXC/UtopiaDaily/internal/common"
	"github.com/UtopiaXC/UtopiaDaily/internal/module"
	"github.com/UtopiaXC/UtopiaDaily/internal/repository"
	"github.com/UtopiaXC/UtopiaDaily/internal/scheduler"
)

type ScheduleJobDTO struct {
	ID            uint       `json:"id"`
	ModuleID      string     `json:"module_id"`
	ModuleKind    string     `json:"module_kind"`
	ModuleName    string     `json:"module_name"`
	ModuleEnabled bool       `json:"module_enabled"`
	TaskKey       string     `json:"task_key"`
	NameKey       string     `json:"name_key"`
	DescKey       string     `json:"description_key"`
	CronExpr      string     `json:"cron_expr"`
	Enabled       bool       `json:"enabled"`
	LastRunAt     *time.Time `json:"last_run_at,omitempty"`
	LastStatus    string     `json:"last_status"`
	LastError     string     `json:"last_error,omitempty"`
}

type TaskDispatcher struct {
	scraperSvc *ScraperService
	pushSvc    *PushService
}

func NewTaskDispatcher(scraperSvc *ScraperService, pushSvc *PushService) *TaskDispatcher {
	return &TaskDispatcher{
		scraperSvc: scraperSvc,
		pushSvc:    pushSvc,
	}
}

func (d *TaskDispatcher) ExecuteModuleTask(ctx context.Context, kind, moduleID, taskKey string) error {
	switch kind {
	case module.KindScraper:
		_, err := d.scraperSvc.RunTask(ctx, moduleID, taskKey)
		return err
	case module.KindPush:
		return d.pushSvc.ExecuteModulePushTask(ctx, moduleID, taskKey)
	default:
		return common.NewAPIError(http.StatusBadRequest, common.ErrScheduleTaskNotFound)
	}
}

func (d *TaskDispatcher) ExecuteUserSubscription(ctx context.Context, subscriptionID uint) error {
	return d.pushSvc.ExecuteUserSubscription(ctx, subscriptionID)
}

type ScheduleService struct {
	modRepo    *repository.ModuleRepository
	dispatcher *TaskDispatcher
	onReload   func()
}

func NewScheduleService(modRepo *repository.ModuleRepository, dispatcher *TaskDispatcher) *ScheduleService {
	return &ScheduleService{
		modRepo:    modRepo,
		dispatcher: dispatcher,
	}
}

func (s *ScheduleService) SetOnReload(fn func()) {
	s.onReload = fn
}

func (s *ScheduleService) ListJobs() ([]ScheduleJobDTO, error) {
	jobs, err := s.modRepo.ListScheduleJobs()
	if err != nil {
		return nil, common.NewAPIError(http.StatusInternalServerError, common.ErrDatabaseError)
	}
	out := make([]ScheduleJobDTO, 0, len(jobs))
	for _, j := range jobs {
		dto := ScheduleJobDTO{
			ID:         j.ID,
			ModuleID:   j.ModuleID,
			TaskKey:    j.TaskKey,
			NameKey:    j.NameKey,
			DescKey:    j.DescriptionKey,
			CronExpr:   j.CronExpr,
			Enabled:    j.Enabled,
			LastRunAt:  j.LastRunAt,
			LastStatus: j.LastStatus,
			LastError:  j.LastError,
		}
		if j.Module != nil {
			dto.ModuleKind = j.Module.Kind
			dto.ModuleName = j.Module.NameKey
			dto.ModuleEnabled = j.Module.Enabled
		}
		out = append(out, dto)
	}
	return out, nil
}

func (s *ScheduleService) UpdateJob(id uint, cronExpr string, enabled bool) error {
	if _, err := s.modRepo.GetScheduleJob(id); err != nil {
		return common.NewAPIError(http.StatusNotFound, common.ErrScheduleTaskNotFound)
	}
	if err := scheduler.ValidateCron(cronExpr); err != nil {
		return common.NewAPIError(http.StatusBadRequest, common.ErrScheduleCronInvalid)
	}
	if err := s.modRepo.UpdateScheduleJob(id, cronExpr, enabled); err != nil {
		return common.NewAPIError(http.StatusInternalServerError, common.ErrDatabaseError)
	}
	if s.onReload != nil {
		s.onReload()
	}
	return nil
}

func (s *ScheduleService) TriggerJob(ctx context.Context, id uint) error {
	job, err := s.modRepo.GetScheduleJob(id)
	if err != nil {
		return common.NewAPIError(http.StatusNotFound, common.ErrScheduleTaskNotFound)
	}
	kind := module.KindScraper
	if job.Module != nil {
		kind = job.Module.Kind
	}
	return s.dispatcher.ExecuteModuleTask(ctx, kind, job.ModuleID, job.TaskKey)
}
