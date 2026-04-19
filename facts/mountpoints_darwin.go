//go:build darwin

package facts

import (
	"fmt"
	"strings"

	"golang.org/x/sys/unix"
)

// MountpointsCollector gathers mounted filesystem facts on macOS.
type MountpointsCollector struct{}

func (m *MountpointsCollector) Collect() (map[string]any, error) {
	result := make(map[string]any)

	var statfs []unix.Statfs_t
	n, err := unix.Getfsstat(statfs, unix.MNT_NOWAIT)
	if err != nil || n == 0 {
		return result, nil
	}

	statfs = make([]unix.Statfs_t, n)
	n, err = unix.Getfsstat(statfs, unix.MNT_NOWAIT)
	if err != nil {
		return result, nil
	}

	for _, stat := range statfs[:n] {
		mountpoint := unix.ByteSliceToString(stat.Mntonname[:])
		fstype := unix.ByteSliceToString(stat.Fstypename[:])
		device := unix.ByteSliceToString(stat.Mntfromname[:])

		if fstype == "devfs" || fstype == "autofs" || fstype == "nullfs" {
			continue
		}

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

	return result, nil
}
