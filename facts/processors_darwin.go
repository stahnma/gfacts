//go:build darwin

package facts

import (
	"runtime"
	"strconv"
	"syscall"
)

// ProcessorsCollector gathers CPU facts on macOS.
type ProcessorsCollector struct{}

func (p *ProcessorsCollector) Collect() (map[string]any, error) {
	result := make(map[string]any)

	result["processors.count"] = runtime.NumCPU()

	if brand, err := syscall.Sysctl("machdep.cpu.brand_string"); err == nil {
		result["processors.models"] = []string{brand}
	}

	// Physical CPU count
	if val, err := syscall.SysctlUint32("hw.physicalcpu"); err == nil {
		result["processors.physicalcount"] = int(val)
	}

	// CPU speed (may not be available on Apple Silicon)
	if val, err := syscall.SysctlUint32("hw.cpufrequency_max"); err == nil && val > 0 {
		mhz := float64(val) / 1_000_000
		result["processors.speed_mhz"] = mhz
		result["processors.speed"] = formatSpeed(mhz)
	}

	return result, nil
}

func formatSpeed(mhz float64) string {
	if mhz >= 1000 {
		return strconv.FormatFloat(mhz/1000, 'f', 2, 64) + " GHz"
	}
	return strconv.FormatFloat(mhz, 'f', 0, 64) + " MHz"
}
