//go:build darwin

package facts

import (
	"fmt"
	"time"

	"golang.org/x/sys/unix"
)

// UptimeCollector gathers system uptime on macOS.
type UptimeCollector struct{}

func (u *UptimeCollector) Collect() (map[string]any, error) {
	result := make(map[string]any)

	raw, err := unix.SysctlRaw("kern.boottime")
	if err != nil || len(raw) < 8 {
		return result, nil
	}

	bootSec := int64(nativeEndianUint64(raw[:8]))
	bootTime := time.Unix(bootSec, 0)
	uptime := time.Since(bootTime)

	totalSecs := int(uptime.Seconds())
	result["uptime.seconds"] = totalSecs
	result["uptime.hours"] = totalSecs / 3600
	result["uptime.days"] = totalSecs / 86400
	result["uptime.uptime"] = humanUptime(totalSecs)

	return result, nil
}

func humanUptime(secs int) string {
	days := secs / 86400
	hours := (secs % 86400) / 3600
	mins := (secs % 3600) / 60

	switch {
	case days > 0:
		return fmt.Sprintf("%d days, %d hours, %d minutes", days, hours, mins)
	case hours > 0:
		return fmt.Sprintf("%d hours, %d minutes", hours, mins)
	default:
		return fmt.Sprintf("%d minutes", mins)
	}
}
