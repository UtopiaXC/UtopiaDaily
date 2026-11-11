package service

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/UtopiaXC/UtopiaDaily/internal/common"
	"github.com/UtopiaXC/UtopiaDaily/internal/event"
	"github.com/UtopiaXC/UtopiaDaily/internal/model"
	"github.com/UtopiaXC/UtopiaDaily/internal/module"
	"github.com/UtopiaXC/UtopiaDaily/internal/module/external"
	"github.com/UtopiaXC/UtopiaDaily/internal/module/registry"
	"github.com/UtopiaXC/UtopiaDaily/internal/pkg/mailer"
	"github.com/UtopiaXC/UtopiaDaily/internal/repository"
	"github.com/UtopiaXC/UtopiaDaily/internal/scheduler"
)

type UserPushSubscriptionDTO struct {
	ID            uint       `json:"id"`
	PushModuleID  string     `json:"push_module_id"`
	Enabled       bool       `json:"enabled"`
	Target        string     `json:"target"`
	CronExpr      string     `json:"cron_expr"`
	DigestTitle   string     `json:"digest_title"`
	MaxItems      int        `json:"max_items"`
	LookbackHours int        `json:"lookback_hours"`
	SourceIDs     []string   `json:"source_ids"`
	LastPushedAt  *time.Time `json:"last_pushed_at,omitempty"`
}

type PushService struct {
	reg         *registry.Registry
	modRepo     *repository.ModuleRepository
	pushRepo    *repository.PushRepository
	scraperRepo *repository.ScraperRepository
	cfgRepo     *repository.ConfigRepository
	userRepo    *repository.UserRepository
	bus         *event.Bus
	onReload    func()
}

func NewPushService(
	reg *registry.Registry,
	modRepo *repository.ModuleRepository,
	pushRepo *repository.PushRepository,
	scraperRepo *repository.ScraperRepository,
	cfgRepo *repository.ConfigRepository,
	userRepo *repository.UserRepository,
	bus *event.Bus,
) *PushService {
	return &PushService{
		reg:         reg,
		modRepo:     modRepo,
		pushRepo:    pushRepo,
		scraperRepo: scraperRepo,
		cfgRepo:     cfgRepo,
		userRepo:    userRepo,
		bus:         bus,
	}
}

func (s *PushService) SetOnScheduleReload(fn func()) {
	s.onReload = fn
}

func (s *PushService) ListModules() ([]ModuleSummaryView, error) {
	list, err := s.modRepo.ListByKind(module.KindPush)
	if err != nil {
		return nil, common.NewAPIError(http.StatusInternalServerError, common.ErrDatabaseError)
	}
	return buildModuleSummaryList(s.reg, list), nil
}

func (s *PushService) ReloadAll() error {
	return reloadAllModulesCommon(s.reg, s.modRepo, s.bus, s.onReload)
}

func (s *PushService) PrepareEnv(ctx context.Context, id string) (string, error) {
	return prepareEnvCommon(ctx, s.reg, s.modRepo, s.bus, module.KindPush, id)
}

func (s *PushService) GetModuleDetail(id string) (*ModuleDetailView, error) {
	modRec, err := s.modRepo.GetByID(id)
	if err != nil || modRec.Kind != module.KindPush {
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

func (s *PushService) EnableModule(ctx context.Context, id string) error {
	return enableModuleCommon(ctx, s.reg, s.modRepo, s.bus, s.onReload, module.KindPush, id)
}

func (s *PushService) DisableModule(id string) error {
	return disableModuleCommon(s.modRepo, s.bus, s.onReload, module.KindPush, id)
}

func (s *PushService) SaveConfig(id string, payload map[string]any) error {
	contract, ok := s.reg.GetPusher(id)
	if !ok {
		return common.NewAPIError(http.StatusNotFound, common.ErrModuleNotFound)
	}
	if err := s.modRepo.SaveConfigValues(id, contract.ConfigSchema(), payload); err != nil {
		return common.NewAPIError(http.StatusInternalServerError, common.ErrDatabaseError)
	}
	return nil
}

func (s *PushService) TestConfig(ctx context.Context, id string, overrides map[string]any) (module.TestResult, error) {
	contract, ok := s.reg.GetPusher(id)
	if !ok {
		return module.TestResult{}, common.NewAPIError(http.StatusNotFound, common.ErrModuleNotFound)
	}
	rows, _ := s.modRepo.GetConfigRows(id)
	sysMap, _ := s.cfgRepo.GetMap()
	acc := repository.NewMemoryConfigAccessor(contract.ConfigSchema(), overrides, rows, sysMap)
	return contract.TestConfig(ctx, acc), nil
}

func (s *PushService) ExecuteModulePushTask(ctx context.Context, id, taskKey string) error {
	contract, ok := s.reg.GetPusher(id)
	if !ok {
		return common.NewAPIError(http.StatusNotFound, common.ErrModuleNotFound)
	}
	modRec, err := s.modRepo.GetByID(id)
	if err != nil {
		return common.NewAPIError(http.StatusNotFound, common.ErrModuleNotFound)
	}
	if meta := contract.Meta(); meta.Source == module.SourceExternal && !modRec.EnvReady && external.HasRequirements(meta.EntryDir) {
		if _, err := s.PrepareEnv(ctx, id); err != nil {
			_ = s.modRepo.RecordScheduleRun(id, taskKey, "failed", err.Error())
			return err
		}
	}

	rows, _ := s.modRepo.GetConfigRows(id)
	sysMap, _ := s.cfgRepo.GetMap()
	acc := repository.NewDBConfigAccessor(contract.ConfigSchema(), rows, sysMap)

	recipient := strings.TrimSpace(acc.GetString("default_recipient"))
	logRecipient := recipient
	if logRecipient == "" {
		logRecipient = "-"
	}
	maxItems := acc.GetInt("max_items", 20)
	if maxItems <= 0 {
		maxItems = 20
	}
	results, err := s.scraperRepo.QueryRecent(1, maxItems, nil)
	if err == nil && len(results) == 0 {
		results, _ = s.scraperRepo.QueryRecent(3, maxItems, nil)
	}
	articles := toDigestArticles(results)

	payload := module.PushPayload{
		Title:     acc.GetString("digest_title"),
		Recipient: recipient,
		Articles:  articles,
	}

	if err := contract.Push(ctx, taskKey, acc, payload); err != nil {
		_ = s.modRepo.RecordScheduleRun(id, taskKey, "failed", err.Error())
		_ = s.pushRepo.CreateDeliveryLog(&model.PushDeliveryLog{
			PushModuleID: id,
			Recipient:    logRecipient,
			Status:       "failed",
			ItemCount:    len(articles),
			ErrorCode:    common.ErrPushDeliveryFailed,
			Details:      err.Error(),
		})
		if s.bus != nil {
			s.bus.Emit(model.EventLevelCritical, model.EventCategoryPush, "event.push.failed", map[string]any{
				"module":    id,
				"recipient": logRecipient,
			}, map[string]any{"error": err.Error()}, "push", id)
		}
		return common.NewAPIError(http.StatusBadGateway, common.ErrPushDeliveryFailed, map[string]any{"detail": err.Error()})
	}

	_ = s.modRepo.RecordScheduleRun(id, taskKey, "success", "")
	_ = s.pushRepo.CreateDeliveryLog(&model.PushDeliveryLog{
		PushModuleID: id,
		Recipient:    logRecipient,
		Status:       "success",
		ItemCount:    len(articles),
	})
	if s.bus != nil {
		s.bus.Emit(model.EventLevelNormal, model.EventCategoryPush, "event.push.delivered", map[string]any{
			"module":    id,
			"recipient": logRecipient,
		}, map[string]any{"item_count": len(articles)}, "push", id)
	}
	return nil
}

func (s *PushService) ListUserSubscriptions(userID uint) ([]UserPushSubscriptionDTO, error) {
	subs, err := s.pushRepo.ListByUser(userID)
	if err != nil {
		return nil, common.NewAPIError(http.StatusInternalServerError, common.ErrDatabaseError)
	}
	out := make([]UserPushSubscriptionDTO, 0, len(subs))
	for _, sub := range subs {
		srcIDs := make([]string, 0, len(sub.Sources))
		for _, src := range sub.Sources {
			srcIDs = append(srcIDs, src.ScraperModuleID)
		}
		out = append(out, UserPushSubscriptionDTO{
			ID:            sub.ID,
			PushModuleID:  sub.PushModuleID,
			Enabled:       sub.Enabled,
			Target:        sub.Target,
			CronExpr:      sub.CronExpr,
			DigestTitle:   sub.DigestTitle,
			MaxItems:      sub.MaxItems,
			LookbackHours: sub.LookbackHours,
			SourceIDs:     srcIDs,
			LastPushedAt:  sub.LastPushedAt,
		})
	}
	return out, nil
}

func (s *PushService) UpsertUserSubscription(userID uint, dto UserPushSubscriptionDTO) (UserPushSubscriptionDTO, error) {
	modRec, err := s.modRepo.GetByID(dto.PushModuleID)
	if err != nil || modRec.Kind != module.KindPush {
		return UserPushSubscriptionDTO{}, common.NewAPIError(http.StatusNotFound, common.ErrModuleNotFound)
	}
	cronExpr := strings.TrimSpace(dto.CronExpr)
	if cronExpr == "" {
		cronExpr = "0 8 * * *"
	}
	if err := scheduler.ValidateCron(cronExpr); err != nil {
		return UserPushSubscriptionDTO{}, common.NewAPIError(http.StatusBadRequest, common.ErrScheduleCronInvalid)
	}
	maxItems := dto.MaxItems
	if maxItems <= 0 {
		maxItems = 25
	}
	lookbackHours := dto.LookbackHours
	if lookbackHours <= 0 {
		lookbackHours = 24
	}

	sub := &model.UserPushSubscription{
		UserID:        userID,
		PushModuleID:  dto.PushModuleID,
		Enabled:       dto.Enabled,
		Target:        strings.TrimSpace(dto.Target),
		CronExpr:      cronExpr,
		DigestTitle:   strings.TrimSpace(dto.DigestTitle),
		MaxItems:      maxItems,
		LookbackHours: lookbackHours,
	}
	if err := s.pushRepo.UpsertSubscription(sub, dto.SourceIDs); err != nil {
		return UserPushSubscriptionDTO{}, common.SanitizeDBError(err)
	}
	if s.onReload != nil {
		s.onReload()
	}

	fresh, err := s.pushRepo.GetByID(sub.ID)
	if err != nil {
		return dto, nil
	}
	srcIDs := make([]string, 0, len(fresh.Sources))
	for _, src := range fresh.Sources {
		srcIDs = append(srcIDs, src.ScraperModuleID)
	}
	return UserPushSubscriptionDTO{
		ID:            fresh.ID,
		PushModuleID:  fresh.PushModuleID,
		Enabled:       fresh.Enabled,
		Target:        fresh.Target,
		CronExpr:      fresh.CronExpr,
		DigestTitle:   fresh.DigestTitle,
		MaxItems:      fresh.MaxItems,
		LookbackHours: fresh.LookbackHours,
		SourceIDs:     srcIDs,
		LastPushedAt:  fresh.LastPushedAt,
	}, nil
}

func (s *PushService) DeleteUserSubscription(userID, subID uint) error {
	if err := s.pushRepo.DeleteSubscription(userID, subID); err != nil {
		return common.NewAPIError(http.StatusInternalServerError, common.ErrDatabaseError)
	}
	if s.onReload != nil {
		s.onReload()
	}
	return nil
}

func (s *PushService) ExecuteUserSubscription(ctx context.Context, subID uint) error {
	sub, err := s.pushRepo.GetByID(subID)
	if err != nil {
		return common.NewAPIError(http.StatusNotFound, common.ErrSubscriptionNotFound)
	}
	user, err := s.userRepo.FindByID(sub.UserID)
	if err != nil {
		return common.NewAPIError(http.StatusNotFound, common.ErrAuthUserNotFound)
	}
	return s.deliverSubscription(ctx, user, sub)
}

func (s *PushService) TestUserSubscription(ctx context.Context, user *model.User, pushModuleID, targetOverride string) (int, error) {
	sub, err := s.pushRepo.GetByUserAndModule(user.ID, pushModuleID)
	if err != nil {
		sub = &model.UserPushSubscription{
			UserID:        user.ID,
			PushModuleID:  pushModuleID,
			MaxItems:      20,
			LookbackHours: 48,
			DigestTitle:   "Daily Digest",
		}
	}
	if strings.TrimSpace(targetOverride) != "" {
		sub.Target = strings.TrimSpace(targetOverride)
	}
	if err := s.deliverSubscription(ctx, user, sub); err != nil {
		return 0, err
	}
	return sub.MaxItems, nil
}

func (s *PushService) deliverSubscription(ctx context.Context, user *model.User, sub *model.UserPushSubscription) error {
	contract, ok := s.reg.GetPusher(sub.PushModuleID)
	if !ok {
		return common.NewAPIError(http.StatusNotFound, common.ErrModuleNotFound)
	}
	modRec, err := s.modRepo.GetByID(sub.PushModuleID)
	if err != nil || !modRec.Enabled {
		return common.NewAPIError(http.StatusBadRequest, common.ErrModuleNotEnabled)
	}
	if meta := contract.Meta(); meta.Source == module.SourceExternal && !modRec.EnvReady && external.HasRequirements(meta.EntryDir) {
		if _, err := s.PrepareEnv(ctx, sub.PushModuleID); err != nil {
			return err
		}
	}

	target := strings.TrimSpace(sub.Target)
	if target == "" && user.Email != nil {
		target = strings.TrimSpace(*user.Email)
	}
	if target == "" {
		return common.NewAPIError(http.StatusBadRequest, common.ErrUserEmailRequired)
	}

	lookback := sub.LookbackHours
	if lookback <= 0 {
		lookback = 24
	}
	maxItems := sub.MaxItems
	if maxItems <= 0 {
		maxItems = 25
	}
	srcIDs := make([]string, 0, len(sub.Sources))
	for _, src := range sub.Sources {
		srcIDs = append(srcIDs, src.ScraperModuleID)
	}

	since := time.Now().UTC().Add(-time.Duration(lookback) * time.Hour)
	results, err := s.scraperRepo.QuerySince(since, maxItems, srcIDs)
	if err != nil {
		return common.NewAPIError(http.StatusInternalServerError, common.ErrDatabaseError)
	}
	articles := toDigestArticles(results)

	rows, _ := s.modRepo.GetConfigRows(sub.PushModuleID)
	sysMap, _ := s.cfgRepo.GetMap()
	acc := repository.NewDBConfigAccessor(contract.ConfigSchema(), rows, sysMap)

	title := strings.TrimSpace(sub.DigestTitle)
	if title == "" {
		title = acc.GetString("digest_title")
	}

	payload := module.PushPayload{
		Title:     title,
		Recipient: target,
		Articles:  articles,
	}

	uid := user.ID
	if err := contract.Push(ctx, "user_digest", acc, payload); err != nil {
		_ = s.pushRepo.CreateDeliveryLog(&model.PushDeliveryLog{
			UserID:       &uid,
			PushModuleID: sub.PushModuleID,
			Recipient:    target,
			Status:       "failed",
			ItemCount:    len(articles),
			ErrorCode:    common.ErrPushDeliveryFailed,
			Details:      err.Error(),
		})
		if s.bus != nil {
			s.bus.Emit(model.EventLevelWarning, model.EventCategoryPush, "event.push.failed", map[string]any{
				"module":    sub.PushModuleID,
				"recipient": target,
			}, map[string]any{"error": err.Error()}, "user", user.Username)
		}
		return common.NewAPIError(http.StatusBadGateway, common.ErrPushDeliveryFailed, map[string]any{"detail": err.Error()})
	}

	if sub.ID > 0 {
		_ = s.pushRepo.MarkPushed(sub.ID)
	}
	_ = s.pushRepo.CreateDeliveryLog(&model.PushDeliveryLog{
		UserID:       &uid,
		PushModuleID: sub.PushModuleID,
		Recipient:    target,
		Status:       "success",
		ItemCount:    len(articles),
	})
	if s.bus != nil {
		s.bus.Emit(model.EventLevelNormal, model.EventCategoryPush, "event.push.delivered", map[string]any{
			"module":    sub.PushModuleID,
			"recipient": target,
		}, map[string]any{"item_count": len(articles)}, "user", user.Username)
	}
	return nil
}

func (s *PushService) ListDeliveryLogs(limit int) ([]model.PushDeliveryLog, error) {
	logs, err := s.pushRepo.ListDeliveryLogs(limit)
	if err != nil {
		return nil, common.NewAPIError(http.StatusInternalServerError, common.ErrDatabaseError)
	}
	return logs, nil
}

func toDigestArticles(results []model.ScraperResult) []mailer.DigestArticle {
	articles := make([]mailer.DigestArticle, 0, len(results))
	for _, r := range results {
		articles = append(articles, mailer.DigestArticle{
			Title:       r.Title,
			Summary:     r.Summary,
			SourceName:  r.Source,
			Author:      r.Author,
			ExternalURL: r.FromURL,
			PublishedAt: r.DatetimeReleased.Format("2006-01-02 15:04"),
		})
	}
	return articles
}
