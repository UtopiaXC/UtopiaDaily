package scheduler

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"github.com/UtopiaXC/UtopiaDaily/internal/repository"
	"github.com/robfig/cron/v3"
)

type JobExecutor interface {
	ExecuteModuleTask(ctx context.Context, kind, moduleID, taskKey string) error
	ExecuteUserSubscription(ctx context.Context, subscriptionID uint) error
}

var cronParser = cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor)

func ValidateCron(expr string) error {
	_, err := cronParser.Parse(expr)
	return err
}

type Scheduler struct {
	mu         sync.Mutex
	c          *cron.Cron
	moduleRepo *repository.ModuleRepository
	pushRepo   *repository.PushRepository
	executor   JobExecutor
	running    map[string]bool
	runningMu  sync.Mutex
	started    bool
}

func NewScheduler(moduleRepo *repository.ModuleRepository, pushRepo *repository.PushRepository, executor JobExecutor) *Scheduler {
	return &Scheduler{
		c:          cron.New(cron.WithParser(cronParser)),
		moduleRepo: moduleRepo,
		pushRepo:   pushRepo,
		executor:   executor,
		running:    make(map[string]bool),
	}
}

func (s *Scheduler) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.started {
		return nil
	}
	if err := s.rebuildLocked(); err != nil {
		return err
	}
	s.c.Start()
	s.started = true
	slog.Info("Task scheduler started")
	return nil
}

func (s *Scheduler) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.started {
		return
	}
	ctx := s.c.Stop()
	<-ctx.Done()
	s.started = false
	slog.Info("Task scheduler stopped")
}

func (s *Scheduler) Reload() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.rebuildLocked()
}

func (s *Scheduler) rebuildLocked() error {
	for _, entry := range s.c.Entries() {
		s.c.Remove(entry.ID)
	}

	jobs, err := s.moduleRepo.ListScheduleJobs()
	if err != nil {
		return err
	}
	for _, j := range jobs {
		if !j.Enabled || j.Module == nil || !j.Module.Enabled {
			continue
		}
		modID := j.ModuleID
		taskKey := j.TaskKey
		kind := j.Module.Kind
		key := fmt.Sprintf("mod:%s:%s", modID, taskKey)

		_, addErr := s.c.AddFunc(j.CronExpr, func() {
			s.runGuarded(key, func() {
				if s.executor == nil {
					return
				}
				if err := s.executor.ExecuteModuleTask(context.Background(), kind, modID, taskKey); err != nil {
					slog.Error("Scheduled module task failed", "module_id", modID, "task_key", taskKey, "error", err)
				}
			})
		})
		if addErr != nil {
			slog.Warn("Invalid schedule cron skipped", "module_id", modID, "task_key", taskKey, "cron", j.CronExpr, "error", addErr)
		}
	}

	subs, err := s.pushRepo.ListEnabledSubscriptions()
	if err == nil {
		mods, _ := s.moduleRepo.ListByKind("push")
		enabledPush := make(map[string]bool, len(mods))
		for _, m := range mods {
			if m.Enabled {
				enabledPush[m.ID] = true
			}
		}
		for _, sub := range subs {
			if !enabledPush[sub.PushModuleID] {
				continue
			}
			subID := sub.ID
			key := fmt.Sprintf("sub:%d", subID)
			_, addErr := s.c.AddFunc(sub.CronExpr, func() {
				s.runGuarded(key, func() {
					if s.executor == nil {
						return
					}
					if err := s.executor.ExecuteUserSubscription(context.Background(), subID); err != nil {
						slog.Error("Scheduled user push failed", "subscription_id", subID, "error", err)
					}
				})
			})
			if addErr != nil {
				slog.Warn("Invalid subscription cron skipped", "subscription_id", subID, "cron", sub.CronExpr, "error", addErr)
			}
		}
	}

	return nil
}

func (s *Scheduler) runGuarded(key string, fn func()) {
	s.runningMu.Lock()
	if s.running[key] {
		s.runningMu.Unlock()
		return
	}
	s.running[key] = true
	s.runningMu.Unlock()

	defer func() {
		s.runningMu.Lock()
		delete(s.running, key)
		s.runningMu.Unlock()
	}()

	fn()
}
