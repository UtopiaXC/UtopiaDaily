package runner

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os/exec"
	"time"
)

type TaskInput struct {
	TaskID string                 `json:"task_id"`
	Config map[string]interface{} `json:"config"`
}

type TaskOutput struct {
	Status  string          `json:"status"`
	Data    json.RawMessage `json:"data"`
	Message string          `json:"message,omitempty"`
}

type LocalPluginRunner struct {
	DefaultTimeout time.Duration
}

func NewLocalPluginRunner(defaultTimeout time.Duration) *LocalPluginRunner {
	if defaultTimeout <= 0 {
		defaultTimeout = 120 * time.Second
	}
	return &LocalPluginRunner{DefaultTimeout: defaultTimeout}
}

func (r *LocalPluginRunner) ExecutePlugin(
	ctx context.Context,
	workingDir string,
	command []string,
	input TaskInput,
) (*TaskOutput, error) {
	if len(command) == 0 {
		return nil, errors.New("command cannot be empty")
	}

	timeoutCtx, cancel := context.WithTimeout(ctx, r.DefaultTimeout)
	defer cancel()

	cmd := exec.CommandContext(timeoutCtx, command[0], command[1:]...)
	cmd.Dir = workingDir

	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf

	stdinPipe, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("create stdin pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("start plugin process: %w", err)
	}

	go func() {
		defer func() { _ = stdinPipe.Close() }()
		_ = json.NewEncoder(stdinPipe).Encode(input)
	}()

	waitErr := cmd.Wait()

	if stderrBuf.Len() > 0 {
		slog.Info("Plugin stderr output", "dir", workingDir, "stderr", stderrBuf.String())
	}

	if waitErr != nil {
		if timeoutCtx.Err() == context.DeadlineExceeded {
			return nil, fmt.Errorf("plugin execution timed out after %v", r.DefaultTimeout)
		}
		return nil, fmt.Errorf("plugin exited with error: %w, stderr: %s", waitErr, stderrBuf.String())
	}

	var output TaskOutput
	if err := json.Unmarshal(stdoutBuf.Bytes(), &output); err != nil {
		return nil, fmt.Errorf("parse plugin output JSON: %w (raw: %s)", err, stdoutBuf.String())
	}

	return &output, nil
}
