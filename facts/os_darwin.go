//go:build darwin

package facts

import (
	"runtime"
	"strings"
	"syscall"

	"golang.org/x/sys/unix"
)

// OSCollector gathers operating system facts on macOS.
type OSCollector struct{}

func (o *OSCollector) Collect() (map[string]any, error) {
	result := make(map[string]any)

	result["os.name"] = "Darwin"
	result["os.family"] = "Darwin"
	result["os.distro.id"] = "macOS"

	var uts unix.Utsname
	if err := unix.Uname(&uts); err == nil {
		result["os.architecture"] = unix.ByteSliceToString(uts.Machine[:])
		result["os.hardware"] = unix.ByteSliceToString(uts.Machine[:])
	}

	if ver, err := syscall.Sysctl("kern.osproductversion"); err == nil {
		result["os.distro.release.full"] = ver
		if major, _, ok := strings.Cut(ver, "."); ok {
			result["os.distro.release.major"] = major
		}
	}

	if build, err := syscall.Sysctl("kern.osversion"); err == nil {
		result["os.build"] = build
	}

	result["os.architecture_go"] = runtime.GOARCH

	return result, nil
}
