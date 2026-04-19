//go:build darwin

package facts

import (
	"strconv"

	"golang.org/x/sys/unix"
)

// MemoryCollector gathers memory facts on macOS.
type MemoryCollector struct{}

func (m *MemoryCollector) Collect() (map[string]any, error) {
	result := make(map[string]any)

	// Total physical memory via sysctl hw.memsize
	if raw, err := unix.SysctlRaw("hw.memsize"); err == nil && len(raw) == 8 {
		total := nativeEndianUint64(raw)
		result["memory.system.total_bytes"] = total
		result["memory.system.total"] = humanBytes(total)
	}

	// Swap info from sysctl vm.swapusage (struct xsw_usage)
	if raw, err := unix.SysctlRaw("vm.swapusage"); err == nil && len(raw) >= 24 {
		swapTotal := nativeEndianUint64(raw[0:8])
		swapUsed := nativeEndianUint64(raw[8:16])
		swapFree := nativeEndianUint64(raw[16:24])
		result["memory.swap.total_bytes"] = swapTotal
		result["memory.swap.total"] = humanBytes(swapTotal)
		result["memory.swap.used_bytes"] = swapUsed
		result["memory.swap.used"] = humanBytes(swapUsed)
		result["memory.swap.free_bytes"] = swapFree
	}

	return result, nil
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

func nativeEndianUint64(b []byte) uint64 {
	return uint64(b[0]) | uint64(b[1])<<8 | uint64(b[2])<<16 | uint64(b[3])<<24 |
		uint64(b[4])<<32 | uint64(b[5])<<40 | uint64(b[6])<<48 | uint64(b[7])<<56
}
