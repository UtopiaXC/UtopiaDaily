package registry

import (
	"os"
	"path/filepath"
	"sync"

	"github.com/UtopiaXC/UtopiaDaily/internal/module"
	"github.com/UtopiaXC/UtopiaDaily/internal/module/external"
	"github.com/UtopiaXC/UtopiaDaily/internal/module/pushers"
	"github.com/UtopiaXC/UtopiaDaily/internal/module/scrapers"
	"github.com/UtopiaXC/UtopiaDaily/internal/pkg/i18n"
	"github.com/UtopiaXC/UtopiaDaily/internal/pkg/runner"
)

type Registry struct {
	mu          sync.RWMutex
	scrapers    map[string]module.ScraperModule
	pushers     map[string]module.PushModule
	pluginsDir  string
	pluginRun   *runner.LocalPluginRunner
	i18nManager *i18n.Assembler
}

func NewRegistry(pluginsDir string, pluginRun *runner.LocalPluginRunner, i18nMgr *i18n.Assembler) *Registry {
	return &Registry{
		scrapers:    make(map[string]module.ScraperModule),
		pushers:     make(map[string]module.PushModule),
		pluginsDir:  pluginsDir,
		pluginRun:   pluginRun,
		i18nManager: i18nMgr,
	}
}

func (r *Registry) RegisterScraper(m module.ScraperModule) {
	r.mu.Lock()
	r.scrapers[m.Meta().ID] = m
	r.mu.Unlock()
}

func (r *Registry) RegisterPusher(m module.PushModule) {
	r.mu.Lock()
	r.pushers[m.Meta().ID] = m
	r.mu.Unlock()
}

func (r *Registry) ReloadAll() error {
	r.mu.Lock()
	r.scrapers = make(map[string]module.ScraperModule)
	r.pushers = make(map[string]module.PushModule)
	r.mu.Unlock()

	r.RegisterScraper(scrapers.NewTelegramChannelScraper())
	r.RegisterScraper(scrapers.NewBilibiliScraper())
	r.RegisterPusher(pushers.NewSMTPPushModule())

	r.scanDir(filepath.Join(r.pluginsDir, "scrapers"), module.KindScraper)
	r.scanDir(filepath.Join(r.pluginsDir, "push"), module.KindPush)

	if r.i18nManager != nil {
		r.i18nManager.ResetModules()
		for _, m := range r.AllModules() {
			locs := m.Locales()
			if len(locs) > 0 {
				r.i18nManager.RegisterModuleLocales(m.Meta().ID, locs)
			}
		}
		r.i18nManager.Rebuild()
	}
	return nil
}

func (r *Registry) scanDir(baseDir, expectedKind string) {
	entries, err := os.ReadDir(baseDir)
	if err != nil {
		return
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		pluginDir := filepath.Join(baseDir, entry.Name())
		manifest, locales, err := external.LoadPluginDir(pluginDir)
		if err != nil || manifest.Kind != expectedKind {
			continue
		}
		if expectedKind == module.KindScraper {
			r.RegisterScraper(external.NewScraperAdapter(pluginDir, manifest, locales, r.pluginRun))
		} else if expectedKind == module.KindPush {
			r.RegisterPusher(external.NewPushAdapter(pluginDir, manifest, locales, r.pluginRun))
		}
	}
}

func (r *Registry) GetScraper(id string) (module.ScraperModule, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	m, ok := r.scrapers[id]
	return m, ok
}

func (r *Registry) GetPusher(id string) (module.PushModule, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	m, ok := r.pushers[id]
	return m, ok
}

func (r *Registry) GetModule(id string) (module.Module, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if m, ok := r.scrapers[id]; ok {
		return m, true
	}
	if m, ok := r.pushers[id]; ok {
		return m, true
	}
	return nil, false
}

func (r *Registry) ListScrapers() []module.ScraperModule {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]module.ScraperModule, 0, len(r.scrapers))
	for _, m := range r.scrapers {
		out = append(out, m)
	}
	return out
}

func (r *Registry) ListPushers() []module.PushModule {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]module.PushModule, 0, len(r.pushers))
	for _, m := range r.pushers {
		out = append(out, m)
	}
	return out
}

func (r *Registry) AllModules() []module.Module {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]module.Module, 0, len(r.scrapers)+len(r.pushers))
	for _, m := range r.scrapers {
		out = append(out, m)
	}
	for _, m := range r.pushers {
		out = append(out, m)
	}
	return out
}
