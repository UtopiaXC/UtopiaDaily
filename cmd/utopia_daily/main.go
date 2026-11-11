package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/UtopiaXC/UtopiaDaily/internal/classification"
	"github.com/UtopiaXC/UtopiaDaily/internal/config"
	"github.com/UtopiaXC/UtopiaDaily/internal/database"
	"github.com/UtopiaXC/UtopiaDaily/internal/event"
	basei18n "github.com/UtopiaXC/UtopiaDaily/internal/i18n"
	"github.com/UtopiaXC/UtopiaDaily/internal/logger"
	"github.com/UtopiaXC/UtopiaDaily/internal/model"
	"github.com/UtopiaXC/UtopiaDaily/internal/module/registry"
	pkgi18n "github.com/UtopiaXC/UtopiaDaily/internal/pkg/i18n"
	"github.com/UtopiaXC/UtopiaDaily/internal/pkg/runner"
	"github.com/UtopiaXC/UtopiaDaily/internal/repository"
	"github.com/UtopiaXC/UtopiaDaily/internal/scheduler"
	"github.com/UtopiaXC/UtopiaDaily/internal/security"
	"github.com/UtopiaXC/UtopiaDaily/internal/service"
	transport "github.com/UtopiaXC/UtopiaDaily/internal/transport/http"
	"github.com/UtopiaXC/UtopiaDaily/web"
)

func main() {
	configPath := flag.String("config", "data/config/config.yaml", "path to yaml configuration file")
	flag.Parse()

	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load configuration: %v\n", err)
		os.Exit(1)
	}
	config.EnsureDirectories(cfg)

	if err := logger.InitLogger(logger.Config{
		Level:          cfg.Logging.Level,
		Dir:            cfg.Logging.Dir,
		MaxSizeMB:      cfg.Logging.MaxSizeMB,
		MaxBackups:     cfg.Logging.MaxBackups,
		MaxAgeDays:     cfg.Logging.MaxAgeDays,
		Compress:       cfg.Logging.Compress,
		PermanentLevel: cfg.Logging.PermanentLevel,
	}); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}

	db, err := database.Open(cfg)
	if err != nil {
		slog.Error("Database initialization failed", "error", err)
		os.Exit(1)
	}

	userRepo := repository.NewUserRepository(db)
	sessionRepo := repository.NewSessionRepository(db)
	roleRepo := repository.NewRoleRepository(db)
	cfgRepo := repository.NewConfigRepository(db)
	eventRepo := repository.NewEventRepository(db)
	modRepo := repository.NewModuleRepository(db)
	scraperRepo := repository.NewScraperRepository(db)
	pushRepo := repository.NewPushRepository(db)

	if lvl, err := cfgRepo.Get("log_level"); err == nil && lvl != "" {
		logger.SetLevel(lvl)
	}
	if permLvl, err := cfgRepo.Get("log_permanent_level"); err == nil && permLvl != "" {
		logger.SetPermanentLevel(permLvl)
	}

	bus := event.NewBus()
	recorder := event.NewDatabaseRecorder(eventRepo, cfgRepo)
	bus.Subscribe(recorder.Handle)

	baseLocalesFS, err := basei18n.BaseLocalesFS()
	if err != nil {
		slog.Error("Embedded locales missing", "error", err)
		os.Exit(1)
	}
	i18nAssembler, err := pkgi18n.NewAssembler(baseLocalesFS, cfg.Storage.CacheDir)
	if err != nil {
		slog.Error("i18n initialization failed", "error", err)
		os.Exit(1)
	}

	jwtMgr := security.NewJWTManager(cfg.Security.JWTSecret, sessionRepo)
	captchaProvider := security.NewAltchaProvider(cfg.Security.JWTSecret)
	attemptsTracker := security.NewLoginAttemptsTracker()

	pluginRunner := runner.NewLocalPluginRunner(120 * time.Second)
	modRegistry := registry.NewRegistry(cfg.Storage.PluginsDir, pluginRunner, i18nAssembler)
	classifier := classification.NewNoopClassifier()

	authSvc := service.NewAuthService(userRepo, roleRepo, cfgRepo, jwtMgr, captchaProvider, attemptsTracker, bus, cfg)
	cfgSvc := service.NewConfigService(cfgRepo, userRepo, scraperRepo, i18nAssembler, classifier, bus, cfg)
	userSvc := service.NewUserService(userRepo, roleRepo, bus)
	scraperSvc := service.NewScraperService(modRegistry, modRepo, scraperRepo, cfgRepo, classifier, bus)
	pushSvc := service.NewPushService(modRegistry, modRepo, pushRepo, scraperRepo, cfgRepo, userRepo, bus)
	dispatcher := service.NewTaskDispatcher(scraperSvc, pushSvc)
	schedSvc := service.NewScheduleService(modRepo, dispatcher)
	newsSvc := service.NewNewspaperService(scraperRepo, cfgSvc)

	if err := scraperSvc.SyncAllModules(); err != nil {
		slog.Error("Initial module sync failed", "error", err)
		os.Exit(1)
	}

	sched := scheduler.NewScheduler(modRepo, pushRepo, dispatcher)
	reloadSched := func() {
		if err := sched.Reload(); err != nil {
			slog.Warn("Scheduler reload failed", "error", err)
		}
	}
	scraperSvc.SetOnScheduleReload(reloadSched)
	pushSvc.SetOnScheduleReload(reloadSched)
	schedSvc.SetOnReload(reloadSched)

	if err := sched.Start(); err != nil {
		slog.Error("Scheduler start failed", "error", err)
		os.Exit(1)
	}

	dashFS, err := web.DashboardFS()
	if err != nil {
		slog.Error("Dashboard assets missing", "error", err)
		os.Exit(1)
	}
	newsFS, err := web.NewspaperFS()
	if err != nil {
		slog.Error("Newspaper assets missing", "error", err)
		os.Exit(1)
	}

	engine := transport.NewRouter(transport.RouterDependencies{
		Mode:        cfg.Server.Mode,
		JWTManager:  jwtMgr,
		Common:      transport.NewCommonController(cfgSvc, i18nAssembler),
		Auth:        transport.NewAuthController(authSvc),
		Dashboard:   transport.NewDashboardController(modRepo, scraperRepo, eventRepo, cfgSvc),
		Config:      transport.NewConfigController(cfgSvc),
		User:        transport.NewUserController(userSvc),
		Scraper:     transport.NewScraperController(scraperSvc),
		Push:        transport.NewPushController(pushSvc, scraperSvc),
		Schedule:    transport.NewScheduleController(schedSvc),
		Newspaper:   transport.NewNewspaperController(newsSvc),
		DashboardFS: dashFS,
		NewspaperFS: newsFS,
	})

	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	srv := &http.Server{
		Addr:              addr,
		Handler:           engine,
		ReadHeaderTimeout: 10 * time.Second,
	}

	bus.Emit(model.EventLevelNormal, model.EventCategorySystem, "event.system.started", nil, map[string]any{
		"address": addr,
		"version": config.Version,
	}, "system", "core")

	slog.Info("Utopia Daily server listening", "address", addr, "version", config.Version)

	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("HTTP server error", "error", err)
			os.Exit(1)
		}
	}()

	stopCh := make(chan os.Signal, 1)
	signal.Notify(stopCh, syscall.SIGINT, syscall.SIGTERM)
	<-stopCh

	slog.Info("Shutting down server...")
	sched.Stop()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("HTTP shutdown error", "error", err)
	}
}
