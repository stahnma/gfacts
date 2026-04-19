//go:build linux

package facts

import "syscall"

// KernelCollector gathers kernel facts on Linux.
type KernelCollector struct{}

func (k *KernelCollector) Collect() (map[string]any, error) {
	result := make(map[string]any)

	var uts syscall.Utsname
	if err := syscall.Uname(&uts); err != nil {
		return result, nil
	}

	result["kernel.name"] = charsToString(uts.Sysname[:])
	result["kernel.release"] = charsToString(uts.Release[:])
	result["kernel.version"] = charsToString(uts.Version[:])

	return result, nil
}
