//go:build darwin

package facts

import "syscall"

// DMICollector gathers hardware model facts on macOS.
// macOS exposes very limited DMI-equivalent data.
type DMICollector struct{}

func (d *DMICollector) Collect() (map[string]any, error) {
	result := make(map[string]any)

	if model, err := syscall.Sysctl("hw.model"); err == nil {
		result["dmi.product.name"] = model
	}
	result["dmi.product.vendor"] = "Apple Inc."

	return result, nil
}
