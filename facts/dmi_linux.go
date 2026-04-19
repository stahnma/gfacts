//go:build linux

package facts

import (
	"os"
	"strings"
)

// DMICollector gathers DMI/SMBIOS facts on Linux.
type DMICollector struct{}

func (d *DMICollector) Collect() (map[string]any, error) {
	result := make(map[string]any)

	dmiFields := map[string]string{
		"dmi.bios.vendor":        "/sys/class/dmi/id/bios_vendor",
		"dmi.bios.version":       "/sys/class/dmi/id/bios_version",
		"dmi.bios.release_date":  "/sys/class/dmi/id/bios_date",
		"dmi.board.manufacturer": "/sys/class/dmi/id/board_vendor",
		"dmi.board.product":      "/sys/class/dmi/id/board_name",
		"dmi.board.serial":       "/sys/class/dmi/id/board_serial",
		"dmi.chassis.type":       "/sys/class/dmi/id/chassis_type",
		"dmi.chassis.vendor":     "/sys/class/dmi/id/chassis_vendor",
		"dmi.product.name":       "/sys/class/dmi/id/product_name",
		"dmi.product.serial":     "/sys/class/dmi/id/product_serial",
		"dmi.product.uuid":       "/sys/class/dmi/id/product_uuid",
		"dmi.product.vendor":     "/sys/class/dmi/id/sys_vendor",
	}

	for key, path := range dmiFields {
		if data, err := os.ReadFile(path); err == nil {
			val := strings.TrimSpace(string(data))
			if val != "" {
				result[key] = val
			}
		}
	}

	return result, nil
}
