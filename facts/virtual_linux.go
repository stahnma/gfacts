//go:build linux

package facts

import (
	"os"
	"strings"
)

// VirtualCollector detects virtualization and containers on Linux.
type VirtualCollector struct{}

func (v *VirtualCollector) Collect() (map[string]any, error) {
	result := make(map[string]any)

	hypervisor := detectHypervisor()
	container := detectContainer()

	if hypervisor != "" {
		result["virtual.is_virtual"] = true
		result["virtual.hypervisor"] = hypervisor
	} else if container != "" {
		result["virtual.is_virtual"] = true
		result["virtual.container"] = container
	} else {
		result["virtual.is_virtual"] = false
	}

	return result, nil
}

func detectHypervisor() string {
	// Check DMI product name
	if data, err := os.ReadFile("/sys/class/dmi/id/product_name"); err == nil {
		name := strings.TrimSpace(strings.ToLower(string(data)))
		switch {
		case strings.Contains(name, "virtualbox"):
			return "virtualbox"
		case strings.Contains(name, "vmware"):
			return "vmware"
		case strings.Contains(name, "kvm"), strings.Contains(name, "qemu"):
			return "kvm"
		case strings.Contains(name, "bochs"):
			return "kvm"
		case name == "hvm domu":
			return "xen"
		}
	}

	// Check sys_vendor
	if data, err := os.ReadFile("/sys/class/dmi/id/sys_vendor"); err == nil {
		vendor := strings.TrimSpace(strings.ToLower(string(data)))
		switch {
		case strings.Contains(vendor, "qemu"):
			return "kvm"
		case strings.Contains(vendor, "vmware"):
			return "vmware"
		case strings.Contains(vendor, "xen"):
			return "xen"
		case strings.Contains(vendor, "parallels"):
			return "parallels"
		case strings.Contains(vendor, "microsoft"):
			return "hyperv"
		}
	}

	// Check /proc/cpuinfo for hypervisor flag
	if data, err := os.ReadFile("/proc/cpuinfo"); err == nil {
		if strings.Contains(string(data), "hypervisor") {
			return "virtual" // generic, detected via CPU flags
		}
	}

	return ""
}

func detectContainer() string {
	// Docker
	if fileExists("/.dockerenv") {
		return "docker"
	}

	// Podman
	if fileExists("/run/.containerenv") {
		return "podman"
	}

	// Check cgroup v1
	if data, err := os.ReadFile("/proc/1/cgroup"); err == nil {
		content := string(data)
		switch {
		case strings.Contains(content, "/docker/"):
			return "docker"
		case strings.Contains(content, "/lxc/"):
			return "lxc"
		case strings.Contains(content, "/kubepods"):
			return "kubernetes"
		}
	}

	// Check cgroup v2
	if data, err := os.ReadFile("/proc/self/mountinfo"); err == nil {
		content := string(data)
		if strings.Contains(content, "/docker/containers/") {
			return "docker"
		}
	}

	// systemd-nspawn
	if data, err := os.ReadFile("/proc/1/environ"); err == nil {
		if strings.Contains(string(data), "container=systemd-nspawn") {
			return "systemd-nspawn"
		}
	}

	return ""
}
