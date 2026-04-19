//go:build linux

package facts

import (
	"os"
	"strconv"
	"strings"
)

// MemoryCollector gathers memory facts on Linux.
type MemoryCollector struct{}

func (m *MemoryCollector) Collect() (map[string]any, error) {
	result := make(map[string]any)

	data, err := os.ReadFile("/proc/meminfo")
	if err != nil {
		return result, nil
	}

	info := parseMeminfo(string(data))

	if total, ok := info["MemTotal"]; ok {
		result["memory.system.total_bytes"] = total
		result["memory.system.total"] = humanBytes(total)
	}
	if free, ok := info["MemFree"]; ok {
		result["memory.system.free_bytes"] = free
	}
	if avail, ok := info["MemAvailable"]; ok {
		result["memory.system.available_bytes"] = avail
		result["memory.system.available"] = humanBytes(avail)
	}
	if total, ok := info["MemTotal"]; ok {
		if avail, ok2 := info["MemAvailable"]; ok2 {
			used := total - avail
			result["memory.system.used_bytes"] = used
			result["memory.system.used"] = humanBytes(used)
		}
	}
	if total, ok := info["SwapTotal"]; ok {
		result["memory.swap.total_bytes"] = total
		result["memory.swap.total"] = humanBytes(total)
	}
	if free, ok := info["SwapFree"]; ok {
		result["memory.swap.free_bytes"] = free
		if total, ok2 := info["SwapTotal"]; ok2 {
			used := total - free
			result["memory.swap.used_bytes"] = used
			result["memory.swap.used"] = humanBytes(used)
		}
	}

	return result, nil
}

// parseMeminfo parses /proc/meminfo and returns values in bytes.
func parseMeminfo(data string) map[string]uint64 {
	result := make(map[string]uint64)
	for _, line := range strings.Split(data, "\n") {
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}
		key := strings.TrimSuffix(parts[0], ":")
		val, err := strconv.ParseUint(parts[1], 10, 64)
		if err != nil {
			continue
		}
		// /proc/meminfo reports in kB
		if len(parts) >= 3 && parts[2] == "kB" {
			val *= 1024
		}
		result[key] = val
	}
	return result
}

func humanBytes(b uint64) string {
	const (
		kib = 1024
		mib = 1024 * kib
		gib = 1024 * mib
		tib = 1024 * gib
	)
	switch {
	case b >= tib:
		return strconv.FormatFloat(float64(b)/float64(tib), 'f', 2, 64) + " TiB"
	case b >= gib:
		return strconv.FormatFloat(float64(b)/float64(gib), 'f', 2, 64) + " GiB"
	case b >= mib:
		return strconv.FormatFloat(float64(b)/float64(mib), 'f', 2, 64) + " MiB"
	case b >= kib:
		return strconv.FormatFloat(float64(b)/float64(kib), 'f', 2, 64) + " KiB"
	default:
		return strconv.FormatUint(b, 10) + " B"
	}
}
