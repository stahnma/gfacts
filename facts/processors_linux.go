//go:build linux

package facts

import (
	"os"
	"runtime"
	"strconv"
	"strings"
)

// ProcessorsCollector gathers CPU facts on Linux.
type ProcessorsCollector struct{}

func (p *ProcessorsCollector) Collect() (map[string]any, error) {
	result := make(map[string]any)

	result["processors.count"] = runtime.NumCPU()

	data, err := os.ReadFile("/proc/cpuinfo")
	if err != nil {
		return result, nil
	}

	models := make(map[string]bool)
	var physicalIDs = make(map[string]bool)
	var speed string

	for _, line := range strings.Split(string(data), "\n") {
		key, val, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		val = strings.TrimSpace(val)

		switch key {
		case "model name", "Processor": // x86 vs ARM
			models[val] = true
		case "cpu MHz":
			speed = val
		case "physical id":
			physicalIDs[val] = true
		}
	}

	modelList := make([]string, 0, len(models))
	for m := range models {
		modelList = append(modelList, m)
	}
	if len(modelList) > 0 {
		result["processors.models"] = modelList
	}

	if speed != "" {
		if mhz, err := strconv.ParseFloat(speed, 64); err == nil {
			result["processors.speed_mhz"] = mhz
			result["processors.speed"] = formatSpeed(mhz)
		}
	}

	if len(physicalIDs) > 0 {
		result["processors.physicalcount"] = len(physicalIDs)
	}

	return result, nil
}

func formatSpeed(mhz float64) string {
	if mhz >= 1000 {
		return strconv.FormatFloat(mhz/1000, 'f', 2, 64) + " GHz"
	}
	return strconv.FormatFloat(mhz, 'f', 0, 64) + " MHz"
}
