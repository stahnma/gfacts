//go:build linux

package facts

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// DisksCollector gathers block device facts on Linux.
type DisksCollector struct{}

func (d *DisksCollector) Collect() (map[string]any, error) {
	result := make(map[string]any)

	entries, err := os.ReadDir("/sys/block")
	if err != nil {
		return result, nil
	}

	for _, entry := range entries {
		name := entry.Name()
		// Skip virtual devices like loop, ram, dm-
		if strings.HasPrefix(name, "loop") || strings.HasPrefix(name, "ram") {
			continue
		}

		prefix := fmt.Sprintf("disks.%s", name)
		base := filepath.Join("/sys/block", name)

		if data, err := os.ReadFile(filepath.Join(base, "size")); err == nil {
			sectors, err := strconv.ParseUint(strings.TrimSpace(string(data)), 10, 64)
			if err == nil {
				bytes := sectors * 512
				result[prefix+".size_bytes"] = bytes
				result[prefix+".size"] = humanBytes(bytes)
			}
		}

		if data, err := os.ReadFile(filepath.Join(base, "device/model")); err == nil {
			result[prefix+".model"] = strings.TrimSpace(string(data))
		}

		if data, err := os.ReadFile(filepath.Join(base, "device/vendor")); err == nil {
			result[prefix+".vendor"] = strings.TrimSpace(string(data))
		}

		if data, err := os.ReadFile(filepath.Join(base, "queue/rotational")); err == nil {
			val := strings.TrimSpace(string(data))
			if val == "0" {
				result[prefix+".type"] = "ssd"
			} else {
				result[prefix+".type"] = "hdd"
			}
		}
	}

	return result, nil
}
