//go:build darwin

package facts

import "syscall"

// VirtualCollector detects virtualization on macOS.
type VirtualCollector struct{}

func (v *VirtualCollector) Collect() (map[string]any, error) {
	result := make(map[string]any)

	// kern.hv_vmm_present is 1 when running inside a hypervisor
	if val, err := syscall.SysctlUint32("kern.hv_vmm_present"); err == nil && val == 1 {
		result["virtual.is_virtual"] = true

		// Try to identify the hypervisor from hw.model
		if model, err := syscall.Sysctl("hw.model"); err == nil {
			switch {
			case contains(model, "VMware"):
				result["virtual.hypervisor"] = "vmware"
			case contains(model, "VirtualBox"):
				result["virtual.hypervisor"] = "virtualbox"
			case contains(model, "Parallels"):
				result["virtual.hypervisor"] = "parallels"
			default:
				result["virtual.hypervisor"] = "virtual"
			}
		} else {
			result["virtual.hypervisor"] = "virtual"
		}
	} else {
		result["virtual.is_virtual"] = false
	}

	return result, nil
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsLower(s, substr))
}

func containsLower(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
