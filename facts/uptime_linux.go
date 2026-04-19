//go:build linux

package facts

import (
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
)

// UptimeCollector gathers system uptime on Linux.
type UptimeCollector struct{}

func (u *UptimeCollector) Collect() (map[string]any, error) {
	result := make(map[string]any)

	data, err := os.ReadFile("/proc/uptime")
	if err != nil {
		return result, nil
	}

	fields := strings.Fields(string(data))
	if len(fields) < 1 {
		return result, nil
	}

	secs, err := strconv.ParseFloat(fields[0], 64)
	if err != nil {
		return result, nil
	}

	totalSecs := int(math.Floor(secs))
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
