package i18n

import (
	"encoding/json"
	"io/fs"
	"maps"
	"os"
	"path/filepath"
	"sync"
)

type LanguageInfo struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

type Assembler struct {
	mu          sync.RWMutex
	baseFS      fs.FS
	cacheDir    string
	languages   []LanguageInfo
	baseLocales map[string]map[string]string
	modLocales  map[string]map[string]map[string]string
	assembled   map[string]map[string]string
}

func NewAssembler(baseFS fs.FS, cacheDir ...string) (*Assembler, error) {
	dir := ""
	if len(cacheDir) > 0 {
		dir = cacheDir[0]
	}
	a := &Assembler{
		baseFS:      baseFS,
		cacheDir:    dir,
		baseLocales: make(map[string]map[string]string),
		modLocales:  make(map[string]map[string]map[string]string),
		assembled:   make(map[string]map[string]string),
	}
	if err := a.loadBase(); err != nil {
		return nil, err
	}
	a.Rebuild()
	return a, nil
}

func (a *Assembler) loadBase() error {
	langData, err := fs.ReadFile(a.baseFS, "languages.json")
	if err != nil {
		return err
	}
	if err := json.Unmarshal(langData, &a.languages); err != nil {
		return err
	}
	for _, lang := range a.languages {
		data, err := fs.ReadFile(a.baseFS, lang.Code+".json")
		if err != nil {
			continue
		}
		var dict map[string]string
		if err := json.Unmarshal(data, &dict); err == nil {
			a.baseLocales[lang.Code] = dict
		}
	}
	return nil
}

func (a *Assembler) ResetModules() {
	a.mu.Lock()
	a.modLocales = make(map[string]map[string]map[string]string)
	a.mu.Unlock()
}

func (a *Assembler) RegisterModuleLocales(moduleID string, locales map[string]map[string]string) {
	a.mu.Lock()
	a.modLocales[moduleID] = locales
	a.mu.Unlock()
}

func (a *Assembler) Rebuild() {
	a.mu.Lock()
	defer a.mu.Unlock()

	merged := make(map[string]map[string]string)
	for _, lang := range a.languages {
		m := make(map[string]string)
		if base, ok := a.baseLocales[lang.Code]; ok {
			maps.Copy(m, base)
		}
		for _, modDicts := range a.modLocales {
			if dict, ok := modDicts[lang.Code]; ok {
				maps.Copy(m, dict)
			}
		}
		merged[lang.Code] = m
	}

	if fallback, ok := merged["en_US"]; ok {
		for code, dict := range merged {
			if code == "en_US" {
				continue
			}
			for k, v := range fallback {
				if _, exists := dict[k]; !exists {
					dict[k] = v
				}
			}
		}
	}

	a.assembled = merged

	if a.cacheDir != "" {
		localesCache := filepath.Join(a.cacheDir, "locales")
		if err := os.MkdirAll(localesCache, 0755); err == nil {
			for code, dict := range merged {
				if data, err := json.MarshalIndent(dict, "", "  "); err == nil {
					_ = os.WriteFile(filepath.Join(localesCache, code+".json"), data, 0644)
				}
			}
		}
	}
}

func (a *Assembler) Languages() []LanguageInfo {
	a.mu.RLock()
	defer a.mu.RUnlock()
	out := make([]LanguageInfo, len(a.languages))
	copy(out, a.languages)
	return out
}

func (a *Assembler) Bundle(locale string) map[string]string {
	a.mu.RLock()
	defer a.mu.RUnlock()

	dict, ok := a.assembled[locale]
	if !ok {
		dict = a.assembled["en_US"]
	}
	out := make(map[string]string, len(dict))
	maps.Copy(out, dict)
	return out
}
