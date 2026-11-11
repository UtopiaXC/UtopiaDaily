package classification

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/UtopiaXC/UtopiaDaily/internal/module/external"
)

type ONNXManager struct {
	runtimeDir string
	modelsDir  string
	mu         sync.Mutex
	setupMu    sync.Mutex
}

func NewONNXManager(cacheDir, modelsDir string) *ONNXManager {
	if strings.TrimSpace(cacheDir) == "" {
		cacheDir = "cache"
	}
	if strings.TrimSpace(modelsDir) == "" {
		modelsDir = filepath.Join("models", "onnx")
	}
	return &ONNXManager{
		runtimeDir: filepath.Join(cacheDir, "onnx_runtime"),
		modelsDir:  modelsDir,
	}
}

func (m *ONNXManager) Provider() string {
	return "onnx"
}

func (m *ONNXManager) EnsureAssets() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if err := os.MkdirAll(m.runtimeDir, 0755); err != nil {
		return err
	}
	if err := os.MkdirAll(m.modelsDir, 0755); err != nil {
		return err
	}

	files := []string{"onnx_worker.py", "requirements.txt", "default_models.json"}
	for _, name := range files {
		data, err := fs.ReadFile(ONNXAssetsFS, "onnx_assets/"+name)
		if err != nil {
			return fmt.Errorf("missing embedded asset %s: %w", name, err)
		}
		target := filepath.Join(m.runtimeDir, name)
		if existing, err := os.ReadFile(target); err == nil && bytes.Equal(existing, data) {
			continue
		}
		if err := os.WriteFile(target, data, 0644); err != nil {
			return fmt.Errorf("failed writing %s: %w", target, err)
		}
	}
	return nil
}

func (m *ONNXManager) EnvStatus() ONNXEnvStatus {
	_ = m.EnsureAssets()
	absRuntime, _ := filepath.Abs(m.runtimeDir)
	absModels, _ := filepath.Abs(m.modelsDir)
	localModels, _ := m.ListLocalModels()
	ready := external.IsVenvReady(m.runtimeDir)
	pyPath := ""
	if ready {
		pyPath = external.VenvPythonPath(absRuntime)
	}
	return ONNXEnvStatus{
		Ready:           ready,
		PythonPath:      pyPath,
		RuntimeDir:      absRuntime,
		ModelsDir:       absModels,
		LocalModelCount: len(localModels),
	}
}

func (m *ONNXManager) PrepareEnv(ctx context.Context) (string, error) {
	return m.PrepareEnvWithProgress(ctx, nil)
}

func (m *ONNXManager) PrepareEnvWithProgress(ctx context.Context, onProgress func(ProgressUpdate)) (string, error) {
	return m.prepareEnvWithProgress(ctx, onProgress, true)
}

func (m *ONNXManager) prepareEnvWithProgress(ctx context.Context, onProgress func(ProgressUpdate), emitReady bool) (string, error) {
	m.setupMu.Lock()
	defer m.setupMu.Unlock()

	if err := m.EnsureAssets(); err != nil {
		return "", err
	}
	if onProgress != nil {
		onProgress(ProgressUpdate{
			Stage:         "preparing_env",
			Message:       "creating python runtime",
			CurrentItem:   "onnx_runtime",
			Percent:       -1,
			Indeterminate: true,
		})
	}
	var lastEmit time.Time
	out, err := external.SetupVenvWithProgress(ctx, m.runtimeDir, func(line string) {
		if onProgress == nil {
			return
		}
		now := time.Now()
		if now.Sub(lastEmit) < 150*time.Millisecond {
			return
		}
		lastEmit = now
		msg := strings.TrimSpace(line)
		if len(msg) > 120 {
			msg = msg[:120]
		}
		onProgress(ProgressUpdate{
			Stage:         "preparing_env",
			Message:       msg,
			CurrentItem:   "onnx_runtime",
			Percent:       -1,
			Indeterminate: true,
		})
	})
	if err == nil && emitReady && onProgress != nil {
		onProgress(ProgressUpdate{
			Stage:         "ready",
			Message:       "environment ready",
			CurrentItem:   "onnx_runtime",
			Percent:       100,
			Indeterminate: false,
		})
	}
	return out, err
}

func (m *ONNXManager) Presets() []PresetModelGroup {
	data, err := fs.ReadFile(ONNXAssetsFS, "onnx_assets/default_models.json")
	if err != nil {
		return nil
	}
	var groups []PresetModelGroup
	if err := json.Unmarshal(data, &groups); err != nil {
		return nil
	}
	return groups
}

func (m *ONNXManager) ListLocalModels() ([]LocalONNXModel, error) {
	if err := os.MkdirAll(m.modelsDir, 0755); err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(m.modelsDir)
	if err != nil {
		return nil, err
	}
	out := make([]LocalONNXModel, 0)
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		dir := filepath.Join(m.modelsDir, entry.Name())
		hasWeight := fileExists(filepath.Join(dir, "model_quantized.onnx")) || fileExists(filepath.Join(dir, "model.onnx"))
		hasTok := fileExists(filepath.Join(dir, "tokenizer.json"))
		ready := hasWeight && hasTok

		label := entry.Name()
		hfRepo := ""
		desc := ""
		sizeAbout := ""
		if infoBytes, err := os.ReadFile(filepath.Join(dir, "info.json")); err == nil {
			var info map[string]string
			if json.Unmarshal(infoBytes, &info) == nil {
				if v := strings.TrimSpace(info["name"]); v != "" {
					label = v
				}
				hfRepo = strings.TrimSpace(info["hf_repo"])
				desc = strings.TrimSpace(info["desc"])
				sizeAbout = strings.TrimSpace(info["size_about"])
			}
		}

		var totalSize int64
		_ = filepath.WalkDir(dir, func(p string, d fs.DirEntry, walkErr error) error {
			if walkErr == nil && !d.IsDir() {
				if fi, stErr := d.Info(); stErr == nil {
					totalSize += fi.Size()
				}
			}
			return nil
		})

		out = append(out, LocalONNXModel{
			Value:     entry.Name(),
			Label:     label,
			HFRepo:    hfRepo,
			Desc:      desc,
			SizeAbout: sizeAbout,
			SizeBytes: totalSize,
			Ready:     ready,
		})
	}
	sort.Slice(out, func(i, j int) bool {
		return strings.ToLower(out[i].Label) < strings.ToLower(out[j].Label)
	})
	return out, nil
}

func (m *ONNXManager) DownloadModel(ctx context.Context, repoID string) error {
	return m.DownloadModelWithProgress(ctx, repoID, nil)
}

func (m *ONNXManager) DownloadModelWithProgress(ctx context.Context, repoID string, onProgress func(ProgressUpdate)) error {
	repoID = strings.TrimSpace(repoID)
	if repoID == "" || !strings.Contains(repoID, "/") || strings.Contains(repoID, "..") {
		return errors.New("invalid HuggingFace repository id")
	}
	if err := m.EnsureAssets(); err != nil {
		return err
	}
	if !external.IsVenvReady(m.runtimeDir) {
		if _, err := m.prepareEnvWithProgress(ctx, onProgress, false); err != nil {
			return fmt.Errorf("failed preparing onnx runtime: %w", err)
		}
		if onProgress != nil {
			onProgress(ProgressUpdate{
				Stage:         "resolving",
				Message:       "checking repository",
				CurrentItem:   repoID,
				Percent:       -1,
				Indeterminate: true,
			})
		}
	}

	absRuntime, _ := filepath.Abs(m.runtimeDir)
	absModels, _ := filepath.Abs(m.modelsDir)
	payload := map[string]any{
		"action":       "download",
		"repo_id":      repoID,
		"models_dir":   absModels,
		"presets_path": filepath.Join(absRuntime, "default_models.json"),
	}
	_, err := m.execWorkerStreaming(ctx, 60*time.Minute, payload, onProgress)
	return err
}

func (m *ONNXManager) DeleteModel(folderName string) error {
	clean := filepath.Clean(strings.TrimSpace(folderName))
	if clean == "" || clean == "." || strings.Contains(clean, "/") || strings.Contains(clean, "\\") || strings.Contains(clean, "..") {
		return errors.New("invalid model name")
	}
	target := filepath.Join(m.modelsDir, clean)
	return os.RemoveAll(target)
}

func (m *ONNXManager) ClassifyItem(ctx context.Context, item InputItem, opts EngineOptions) ([]TagAssignment, error) {
	modelName := strings.TrimSpace(opts.ModelName)
	if modelName == "" || strings.EqualFold(modelName, "none") {
		return nil, errors.New("no ONNX classification model configured")
	}
	cleanName := filepath.Clean(modelName)
	if strings.Contains(cleanName, "/") || strings.Contains(cleanName, "\\") || strings.Contains(cleanName, "..") {
		return nil, errors.New("invalid ONNX model name")
	}
	if err := m.EnsureAssets(); err != nil {
		return nil, err
	}
	if !external.IsVenvReady(m.runtimeDir) {
		return nil, errors.New("ONNX Python runtime environment is not prepared")
	}

	absModels, _ := filepath.Abs(m.modelsDir)
	modelDir := filepath.Join(absModels, cleanName)
	if !fileExists(filepath.Join(modelDir, "tokenizer.json")) {
		return nil, fmt.Errorf("ONNX model %q not found locally", cleanName)
	}

	if len(opts.Profile.Groups) == 0 {
		return nil, errors.New("no classification groups or tags configured")
	}

	text := PrepareClassificationText(item, opts.MaxInputChars, opts.TruncateStrategy)
	if strings.TrimSpace(text) == "" {
		return nil, errors.New("article content is empty")
	}

	conf := opts.Confidence
	if conf <= 0 {
		conf = 60
	}

	payload := map[string]any{
		"action":            "classify",
		"model_dir":         modelDir,
		"model_name":        cleanName,
		"text":              text,
		"confidence":        conf,
		"max_input_chars":   opts.MaxInputChars,
		"truncate_strategy": NormalizeTruncateStrategy(opts.TruncateStrategy),
		"profile":           opts.Profile,
	}
	stdout, err := m.execWorker(ctx, 90*time.Second, payload)
	if err != nil {
		return nil, err
	}

	var parsed struct {
		Status string `json:"status"`
		Tags   []struct {
			TagID      uint    `json:"tag_id"`
			TagName    string  `json:"tag_name"`
			GroupName  string  `json:"group_name"`
			Confidence float64 `json:"confidence"`
			Reason     string  `json:"reason"`
		} `json:"tags"`
		BestCandidate *struct {
			TagID      uint    `json:"tag_id"`
			TagName    string  `json:"tag_name"`
			GroupName  string  `json:"group_name"`
			Confidence float64 `json:"confidence"`
		} `json:"best_candidate"`
	}
	if err := json.Unmarshal(stdout, &parsed); err != nil {
		return nil, fmt.Errorf("invalid onnx worker json output: %w", err)
	}

	out := make([]TagAssignment, 0, len(parsed.Tags))
	for _, t := range parsed.Tags {
		if t.TagName == "" {
			continue
		}
		out = append(out, TagAssignment{
			TagID:      t.TagID,
			TagName:    t.TagName,
			GroupName:  t.GroupName,
			Confidence: t.Confidence,
			Provider:   "onnx",
			Reason:     t.Reason,
		})
	}
	if len(out) == 0 {
		if parsed.BestCandidate != nil && parsed.BestCandidate.TagName != "" {
			return nil, fmt.Errorf("no tags met confidence threshold %.2f (highest candidate: %s %.2f)", float64(conf)/100.0, parsed.BestCandidate.TagName, parsed.BestCandidate.Confidence)
		}
		return nil, errors.New("ONNX model returned 0 matching tags")
	}
	return out, nil
}

type workerProgressStderrWriter struct {
	pending    string
	errLines   []string
	onProgress func(ProgressUpdate)
	mu         sync.Mutex
}

func (w *workerProgressStderrWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.pending += strings.ReplaceAll(string(p), "\r", "\n")
	for {
		idx := strings.IndexByte(w.pending, '\n')
		if idx < 0 {
			break
		}
		line := strings.TrimSpace(w.pending[:idx])
		w.pending = w.pending[idx+1:]
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "PROGRESS_JSON:") {
			if w.onProgress != nil {
				var upd ProgressUpdate
				if json.Unmarshal([]byte(strings.TrimPrefix(line, "PROGRESS_JSON:")), &upd) == nil {
					w.onProgress(upd)
				}
			}
			continue
		}
		w.errLines = append(w.errLines, line)
	}
	return len(p), nil
}

func (w *workerProgressStderrWriter) ErrorOutput() string {
	w.mu.Lock()
	defer w.mu.Unlock()
	if rem := strings.TrimSpace(w.pending); rem != "" && !strings.HasPrefix(rem, "PROGRESS_JSON:") {
		w.errLines = append(w.errLines, rem)
	}
	return strings.TrimSpace(strings.Join(w.errLines, "\n"))
}

func (m *ONNXManager) execWorker(ctx context.Context, timeout time.Duration, payload any) ([]byte, error) {
	return m.execWorkerStreaming(ctx, timeout, payload, nil)
}

func (m *ONNXManager) execWorkerStreaming(ctx context.Context, timeout time.Duration, payload any, onProgress func(ProgressUpdate)) ([]byte, error) {
	tctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	rawIn, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	absRuntime, _ := filepath.Abs(m.runtimeDir)
	cmdArgs := external.ResolveCommand(absRuntime, []string{"python3", "onnx_worker.py"})
	cmd := exec.CommandContext(tctx, cmdArgs[0], cmdArgs[1:]...)
	cmd.Dir = absRuntime
	cmd.Stdin = bytes.NewReader(rawIn)

	var stdout bytes.Buffer
	stderr := &workerProgressStderrWriter{onProgress: onProgress}
	cmd.Stdout = &stdout
	cmd.Stderr = stderr

	if err := cmd.Run(); err != nil {
		errDetail := stderr.ErrorOutput()
		if errDetail == "" {
			errDetail = err.Error()
		}
		return nil, fmt.Errorf("onnx worker failed: %s", errDetail)
	}
	return bytes.TrimSpace(stdout.Bytes()), nil
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}
