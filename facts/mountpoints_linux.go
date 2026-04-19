//go:build linux

package facts

import (
	"fmt"
	"os"
	"strings"
	"syscall"
)

// MountpointsCollector gathers mounted filesystem facts on Linux.
type MountpointsCollector struct{}

func (m *MountpointsCollector) Collect() (map[string]any, error) {
	result := make(map[string]any)

	data, err := os.ReadFile("/proc/mounts")
	if err != nil {
		return result, nil
	}

	for _, line := range strings.Split(string(data), "\n") {
		fields := strings.Fields(line)
		if len(fields) < 3 {
			continue
		}
		device := fields[0]
		mountpoint := fields[1]
		fstype := fields[2]

		// Skip virtual/pseudo filesystems
		if !strings.HasPrefix(device, "/") {
			continue
		}

		// Sanitize mountpoint for use as fact key
		key := mountpoint
		if key == "/" {
			key = "_root"
		} else {
			key = strings.ReplaceAll(strings.TrimPrefix(key, "/"), "/", "_")
		}

		prefix := fmt.Sprintf("mountpoints.%s", key)
		result[prefix+".device"] = device
		result[prefix+".path"] = mountpoint
		result[prefix+".filesystem"] = fstype

		var stat syscall.Statfs_t
		if err := syscall.Statfs(mountpoint, &stat); err == nil {
			total := stat.Blocks * uint64(stat.Bsize)
			free := stat.Bfree * uint64(stat.Bsize)
			avail := stat.Bavail * uint64(stat.Bsize)
			used := total - free

			result[prefix+".size_bytes"] = total
			result[prefix+".size"] = humanBytes(total)
			result[prefix+".used_bytes"] = used
			result[prefix+".used"] = humanBytes(used)
			result[prefix+".available_bytes"] = avail
			result[prefix+".available"] = humanBytes(avail)
			if total > 0 {
				result[prefix+".used_percent"] = fmt.Sprintf("%.1f%%", float64(used)/float64(total)*100)
			}
		}
	}

	return result, nil
}
