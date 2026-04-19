//go:build linux

package facts

func platformCollectors() []Collector {
	return []Collector{
		&OSCollector{},
		&KernelCollector{},
		&NetworkingCollector{},
		&ProcessorsCollector{},
		&MemoryCollector{},
		&DMICollector{},
		&DisksCollector{},
		&MountpointsCollector{},
		&UptimeCollector{},
		&CloudCollector{},
		&VirtualCollector{},
		&SSHCollector{},
	}
}
