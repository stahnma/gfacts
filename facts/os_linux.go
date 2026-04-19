//go:build linux

package facts

import (
	"os"
	"runtime"
	"strings"
	"syscall"
)

// OSCollector gathers operating system facts on Linux.
type OSCollector struct{}

func (o *OSCollector) Collect() (map[string]any, error) {
	result := make(map[string]any)

	var uts syscall.Utsname
	if err := syscall.Uname(&uts); err == nil {
		result["os.architecture"] = charsToString(uts.Machine[:])
		result["os.hardware"] = charsToString(uts.Machine[:])
	}
	result["os.family"] = "Linux"

	osRelease := parseOSRelease()
	if name, ok := osRelease["NAME"]; ok {
		result["os.name"] = name
	}
	if id, ok := osRelease["ID"]; ok {
		result["os.distro.id"] = id
		result["os.family"] = distroFamily(id)
	}
	if pretty, ok := osRelease["PRETTY_NAME"]; ok {
		result["os.distro.description"] = pretty
	}
	if version, ok := osRelease["VERSION_ID"]; ok {
		result["os.distro.release.full"] = version
		if major, _, ok := strings.Cut(version, "."); ok {
			result["os.distro.release.major"] = major
		} else {
			result["os.distro.release.major"] = version
		}
	}
	if codename, ok := osRelease["VERSION_CODENAME"]; ok {
		result["os.distro.codename"] = codename
	}

	result["os.selinux.enabled"] = fileExists("/etc/selinux/config")
	result["os.architecture_go"] = runtime.GOARCH

	return result, nil
}

func parseOSRelease() map[string]string {
	result := make(map[string]string)
	data, err := os.ReadFile("/etc/os-release")
	if err != nil {
		return result
	}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, val, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		val = strings.Trim(val, `"`)
		result[key] = val
	}
	return result
}

func distroFamily(id string) string {
	switch id {
	case "ubuntu", "debian", "linuxmint", "pop", "elementary", "kali", "raspbian":
		return "Debian"
	case "fedora", "rhel", "centos", "rocky", "alma", "amzn", "ol", "scientific":
		return "RedHat"
	case "opensuse", "sles", "suse":
		return "Suse"
	case "arch", "manjaro", "endeavouros":
		return "Archlinux"
	case "alpine":
		return "Alpine"
	case "gentoo":
		return "Gentoo"
	case "slackware":
		return "Slackware"
	default:
		return "Linux"
	}
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func charsToString(chars []int8) string {
	b := make([]byte, 0, len(chars))
	for _, c := range chars {
		if c == 0 {
			break
		}
		b = append(b, byte(c))
	}
	return string(b)
}
