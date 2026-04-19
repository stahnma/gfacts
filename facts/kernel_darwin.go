//go:build darwin

package facts

import "golang.org/x/sys/unix"

// KernelCollector gathers kernel facts on macOS.
type KernelCollector struct{}

func (k *KernelCollector) Collect() (map[string]any, error) {
	result := make(map[string]any)

	var uts unix.Utsname
	if err := unix.Uname(&uts); err != nil {
		return result, nil
	}

	result["kernel.name"] = unix.ByteSliceToString(uts.Sysname[:])
	result["kernel.release"] = unix.ByteSliceToString(uts.Release[:])
	result["kernel.version"] = unix.ByteSliceToString(uts.Version[:])

	return result, nil
}
