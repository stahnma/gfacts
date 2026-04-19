package external

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os/exec"
	"strings"
	"time"
)

func loadExec(path string, timeout time.Duration) (map[string]any, error) {
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, path)
	out, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			stderr := strings.TrimSpace(string(exitErr.Stderr))
			if stderr != "" {
				slog.Warn("executable fact stderr", "path", path, "stderr", stderr)
			}
			return nil, fmt.Errorf("exit code %d", exitErr.ExitCode())
		}
		return nil, err
	}

	// Try JSON first.
	var jsonData map[string]any
	if err := json.Unmarshal(out, &jsonData); err == nil {
		result := make(map[string]any)
		flatten("", jsonData, result)
		return result, nil
	}

	// Fall back to key=value.
	result := make(map[string]any)
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, val, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		result[strings.TrimSpace(key)] = strings.TrimSpace(val)
	}
	return result, nil
}
