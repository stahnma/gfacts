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
		"hardware.bios.vendor":        "/sys/class/dmi/id/bios_vendor",
		"hardware.bios.version":       "/sys/class/dmi/id/bios_version",
		"hardware.bios.release_date":  "/sys/class/dmi/id/bios_date",
		"hardware.board.manufacturer": "/sys/class/dmi/id/board_vendor",
		"hardware.board.product":      "/sys/class/dmi/id/board_name",
		"hardware.board.serial":       "/sys/class/dmi/id/board_serial",
		"hardware.chassis.type":       "/sys/class/dmi/id/chassis_type",
		"hardware.chassis.vendor":     "/sys/class/dmi/id/chassis_vendor",
		"hardware.product.name":       "/sys/class/dmi/id/product_name",
		"hardware.product.serial":     "/sys/class/dmi/id/product_serial",
		"hardware.product.uuid":       "/sys/class/dmi/id/product_uuid",
		"hardware.product.vendor":     "/sys/class/dmi/id/sys_vendor",
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
