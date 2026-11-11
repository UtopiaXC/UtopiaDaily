package external

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
)

func IsPythonCommand(cmd []string) bool {
	if len(cmd) == 0 {
		return false
	}
	base := strings.ToLower(filepath.Base(cmd[0]))
	return base == "python" || base == "python3" || strings.HasPrefix(base, "python3.")
}

func HasRequirements(dir string) bool {
	info, err := os.Stat(filepath.Join(dir, "requirements.txt"))
	return err == nil && !info.IsDir()
}

func VenvPythonPath(dir string) string {
	return filepath.Join(dir, ".venv", "bin", "python")
}

func IsVenvReady(dir string) bool {
	py := VenvPythonPath(dir)
	info, err := os.Stat(py)
	return err == nil && !info.IsDir()
}

func ResolveCommand(dir string, command []string) []string {
	if len(command) == 0 {
		return command
	}
	out := make([]string, len(command))
	copy(out, command)
	if IsPythonCommand(out) && IsVenvReady(dir) {
		absPy, err := filepath.Abs(VenvPythonPath(dir))
		if err == nil {
			out[0] = absPy
		} else {
			out[0] = VenvPythonPath(dir)
		}
	}
	return out
}

type lineNotifyWriter struct {
	buf     bytes.Buffer
	pending string
	onLine  func(string)
	mu      sync.Mutex
}

func (w *lineNotifyWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.buf.Write(p)
	if w.onLine == nil {
		return len(p), nil
	}
	normalized := strings.ReplaceAll(string(p), "\r", "\n")
	w.pending += normalized
	for {
		idx := strings.IndexByte(w.pending, '\n')
		if idx < 0 {
			break
		}
		line := strings.TrimSpace(w.pending[:idx])
		w.pending = w.pending[idx+1:]
		if line != "" {
			w.onLine(line)
		}
	}
	return len(p), nil
}

func (w *lineNotifyWriter) String() string {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.buf.String()
}

func SetupVenv(ctx context.Context, dir string) (string, error) {
	return SetupVenvWithProgress(ctx, dir, nil)
}

func SetupVenvWithProgress(ctx context.Context, dir string, onLine func(string)) (string, error) {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		absDir = dir
	}
	venvDir := filepath.Join(absDir, ".venv")
	reqFile := filepath.Join(absDir, "requirements.txt")
	hasReq := false
	if _, statErr := os.Stat(reqFile); statErr == nil {
		hasReq = true
	}

	out := &lineNotifyWriter{onLine: onLine}
	runStep := func(name string, args ...string) error {
		cmd := exec.CommandContext(ctx, name, args...)
		cmd.Stdout = out
		cmd.Stderr = io.Writer(out)
		return cmd.Run()
	}

	if uvPath, err := exec.LookPath("uv"); err == nil {
		if onLine != nil {
			onLine("uv venv " + venvDir)
		}
		if err := runStep(uvPath, "venv", venvDir); err != nil {
			return out.String(), fmt.Errorf("uv venv failed: %w", err)
		}
		if hasReq {
			if onLine != nil {
				onLine("uv pip install -r requirements.txt")
			}
			pyBin := VenvPythonPath(absDir)
			if err := runStep(uvPath, "pip", "install", "--python", pyBin, "-r", reqFile); err != nil {
				return out.String(), fmt.Errorf("uv pip install failed: %w", err)
			}
		}
		return out.String(), nil
	}

	pyHost, err := exec.LookPath("python3")
	if err != nil {
		pyHost, err = exec.LookPath("python")
		if err != nil {
			return "", errors.New("neither uv nor python3 found in PATH")
		}
	}

	if onLine != nil {
		onLine("python3 -m venv " + venvDir)
	}
	if err := runStep(pyHost, "-m", "venv", venvDir); err != nil {
		return out.String(), fmt.Errorf("python venv failed: %w", err)
	}
	if hasReq {
		if onLine != nil {
			onLine("pip install -r requirements.txt")
		}
		pipBin := filepath.Join(venvDir, "bin", "pip")
		if err := runStep(pipBin, "install", "--progress-bar", "off", "-r", reqFile); err != nil {
			return out.String(), fmt.Errorf("pip install failed: %w", err)
		}
	}
	return out.String(), nil
}
