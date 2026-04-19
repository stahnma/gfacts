//go:build darwin

package facts

import (
	"encoding/json"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

// DMICollector gathers hardware model facts on macOS.
type DMICollector struct{}

var cacheDir string
var cacheFile string

func init() {
	home, err := os.UserHomeDir()
	if err != nil {
		home = os.TempDir()
	}
	cacheDir = filepath.Join(home, ".cache", "gfacts")
	cacheFile = filepath.Join(cacheDir, "hardware_profile.json")
}

func (d *DMICollector) Collect() (map[string]any, error) {
	result := make(map[string]any)

	if model, err := syscall.Sysctl("hw.model"); err == nil {
		result["hardware.product.identifier"] = model
	}
	result["hardware.product.vendor"] = "Apple Inc."

	hw := loadHardwareProfile()
	for k, v := range hw {
		result[k] = v
	}

	return result, nil
}

const cacheTTL = 30 * 24 * time.Hour // 30 days

func loadHardwareProfile() map[string]string {
	// Try cache if it exists and hasn't expired.
	if info, err := os.Stat(cacheFile); err == nil {
		if time.Since(info.ModTime()) < cacheTTL {
			if data, err := os.ReadFile(cacheFile); err == nil {
				var cached map[string]string
				if err := json.Unmarshal(data, &cached); err == nil && len(cached) > 0 {
					return cached
				}
			}
		}
	}

	// Gather hardware info from system_profiler and ioreg.
	hw := runSystemProfiler()
	if hw == nil {
		hw = make(map[string]string)
	}

	if desc := productDescription(); desc != "" {
		hw["hardware.product.description"] = desc
	}

	if len(hw) == 0 {
		return hw
	}

	// Cache for future runs.
	if err := os.MkdirAll(cacheDir, 0755); err == nil {
		if data, err := json.Marshal(hw); err == nil {
			if err := os.WriteFile(cacheFile, data, 0644); err != nil {
				slog.Debug("failed to write hardware cache", "error", err)
			}
		}
	}

	return hw
}

func runSystemProfiler() map[string]string {
	out, err := exec.Command("system_profiler", "SPHardwareDataType").Output()
	if err != nil {
		slog.Debug("system_profiler exec failed", "error", err)
		return nil
	}

	result := make(map[string]string)
	fieldMap := map[string]string{
		"Model Name":       "hardware.product.name",
		"Model Number":     "hardware.product.model_number",
		"Chip":             "hardware.product.chip",
		"Serial Number":    "hardware.product.serial",
		"Hardware UUID":    "hardware.product.uuid",
		"Total Number of Cores": "hardware.product.cores",
	}

	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		for label, key := range fieldMap {
			if strings.HasPrefix(line, label) {
				_, val, ok := strings.Cut(line, ":")
				if ok {
					result[key] = strings.TrimSpace(val)
				}
			}
		}
	}

	return result
}

// productDescription returns the marketing description from ioreg,
// e.g. "MacBook Air (13-inch, M4, 2025)".
func productDescription() string {
	out, err := exec.Command("ioreg", "-r", "-c", "IOPlatformDevice", "-d1").Output()
	if err != nil {
		return ""
	}
	for _, line := range strings.Split(string(out), "\n") {
		if strings.Contains(line, "product-description") {
			// Format: "product-description" = <"MacBook Air (13-inch, M4, 2025)">
			if start := strings.Index(line, `<"`); start != -1 {
				if end := strings.Index(line[start:], `">`); end != -1 {
					return line[start+2 : start+end]
				}
			}
		}
	}
	return ""
}
