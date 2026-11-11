package external

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/UtopiaXC/UtopiaDaily/internal/module"
)

type Manifest struct {
	ID          string                   `json:"id"`
	Kind        string                   `json:"kind"`
	Name        string                   `json:"name,omitempty"`
	Description string                   `json:"description,omitempty"`
	Version     string                   `json:"version"`
	Author      string                   `json:"author"`
	Command     []string                 `json:"command"`
	Configs     []module.ConfigFieldSpec `json:"configs"`
	Tasks       []module.TaskSpec        `json:"tasks"`
}

func LoadPluginDir(dir string) (*Manifest, map[string]map[string]string, error) {
	manifestPath := filepath.Join(dir, "manifest.json")
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return nil, nil, err
	}
	var m Manifest
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, nil, fmt.Errorf("parse manifest %s: %w", manifestPath, err)
	}
	if strings.TrimSpace(m.ID) == "" {
		return nil, nil, errors.New("manifest missing id")
	}
	if m.Kind == "" {
		m.Kind = module.KindScraper
	}
	if m.Kind != module.KindScraper && m.Kind != module.KindPush {
		return nil, nil, fmt.Errorf("invalid manifest kind: %s", m.Kind)
	}
	if len(m.Command) == 0 {
		return nil, nil, errors.New("manifest missing command")
	}
	if m.Version == "" {
		m.Version = module.DefaultVersion
	}

	locales := make(map[string]map[string]string)
	localesDir := filepath.Join(dir, "locales")
	entries, err := os.ReadDir(localesDir)
	if err == nil {
		for _, entry := range entries {
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
				continue
			}
			localeCode := strings.TrimSuffix(entry.Name(), ".json")
			raw, err := os.ReadFile(filepath.Join(localesDir, entry.Name()))
			if err != nil {
				continue
			}
			var dict map[string]string
			if err := json.Unmarshal(raw, &dict); err == nil {
				locales[localeCode] = dict
			}
		}
	}

	return &m, locales, nil
}
