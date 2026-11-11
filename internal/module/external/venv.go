package external

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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

func SetupVenv(ctx context.Context, dir string) (string, error) {
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

	var out bytes.Buffer
	if uvPath, err := exec.LookPath("uv"); err == nil {
		createCmd := exec.CommandContext(ctx, uvPath, "venv", venvDir)
		createCmd.Stdout = &out
		createCmd.Stderr = &out
		if err := createCmd.Run(); err != nil {
			return out.String(), fmt.Errorf("uv venv failed: %w", err)
		}
		if hasReq {
			pyBin := VenvPythonPath(absDir)
			installCmd := exec.CommandContext(ctx, uvPath, "pip", "install", "--python", pyBin, "-r", reqFile)
			installCmd.Stdout = &out
			installCmd.Stderr = &out
			if err := installCmd.Run(); err != nil {
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

	createCmd := exec.CommandContext(ctx, pyHost, "-m", "venv", venvDir)
	createCmd.Stdout = &out
	createCmd.Stderr = &out
	if err := createCmd.Run(); err != nil {
		return out.String(), fmt.Errorf("python venv failed: %w", err)
	}
	if hasReq {
		pipBin := filepath.Join(venvDir, "bin", "pip")
		installCmd := exec.CommandContext(ctx, pipBin, "install", "-r", reqFile)
		installCmd.Stdout = &out
		installCmd.Stderr = &out
		if err := installCmd.Run(); err != nil {
			return out.String(), fmt.Errorf("pip install failed: %w", err)
		}
	}
	return out.String(), nil
}
