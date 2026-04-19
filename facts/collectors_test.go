package facts

import (
	"strings"
	"testing"
)

func TestAllCollectors(t *testing.T) {
	collectors := All()
	if len(collectors) == 0 {
		t.Fatal("All() returned no collectors")
	}
	// On darwin we expect 12 collectors.
	t.Logf("got %d collectors", len(collectors))
}

func TestOSCollector(t *testing.T) {
	c := &OSCollector{}
	facts, err := c.Collect()
	if err != nil {
		t.Fatalf("OSCollector error: %v", err)
	}

	requiredKeys := []string{"os.name", "os.family"}
	for _, k := range requiredKeys {
		if _, ok := facts[k]; !ok {
			t.Errorf("missing expected key %q", k)
		}
	}

	// os.name should be a non-empty string.
	if name, ok := facts["os.name"].(string); !ok || name == "" {
		t.Errorf("os.name should be a non-empty string, got %v", facts["os.name"])
	}
}

func TestKernelCollector(t *testing.T) {
	c := &KernelCollector{}
	facts, err := c.Collect()
	if err != nil {
		t.Fatalf("KernelCollector error: %v", err)
	}

	requiredKeys := []string{"kernel.name", "kernel.release", "kernel.version"}
	for _, k := range requiredKeys {
		if _, ok := facts[k]; !ok {
			t.Errorf("missing expected key %q", k)
		}
	}
}

func TestNetworkingCollector(t *testing.T) {
	c := &NetworkingCollector{}
	facts, err := c.Collect()
	if err != nil {
		t.Fatalf("NetworkingCollector error: %v", err)
	}

	if _, ok := facts["networking.hostname"]; !ok {
		t.Error("missing networking.hostname")
	}

	// Should have at least one interface key.
	hasInterface := false
	for k := range facts {
		if strings.HasPrefix(k, "networking.interfaces.") {
			hasInterface = true
			break
		}
	}
	if !hasInterface {
		t.Log("warning: no interface facts found (may be expected in some environments)")
	}
}

func TestProcessorsCollector(t *testing.T) {
	c := &ProcessorsCollector{}
	facts, err := c.Collect()
	if err != nil {
		t.Fatalf("ProcessorsCollector error: %v", err)
	}

	if _, ok := facts["processors.count"]; !ok {
		t.Error("missing processors.count")
	}

	count, ok := facts["processors.count"].(int)
	if !ok {
		t.Errorf("processors.count should be int, got %T", facts["processors.count"])
	} else if count < 1 {
		t.Errorf("processors.count should be >= 1, got %d", count)
	}
}

func TestMemoryCollector(t *testing.T) {
	c := &MemoryCollector{}
	facts, err := c.Collect()
	if err != nil {
		t.Fatalf("MemoryCollector error: %v", err)
	}

	if _, ok := facts["memory.system.total_bytes"]; !ok {
		t.Error("missing memory.system.total_bytes")
	}
	if _, ok := facts["memory.system.total"]; !ok {
		t.Error("missing memory.system.total")
	}
}

func TestDMICollector(t *testing.T) {
	c := &DMICollector{}
	facts, err := c.Collect()
	if err != nil {
		t.Fatalf("DMICollector error: %v", err)
	}

	if _, ok := facts["dmi.product.name"]; !ok {
		t.Error("missing dmi.product.name")
	}
	if facts["dmi.product.vendor"] != "Apple Inc." {
		t.Errorf("expected Apple Inc., got %v", facts["dmi.product.vendor"])
	}
}

func TestUptimeCollector(t *testing.T) {
	c := &UptimeCollector{}
	facts, err := c.Collect()
	if err != nil {
		t.Fatalf("UptimeCollector error: %v", err)
	}

	if _, ok := facts["uptime.seconds"]; !ok {
		t.Error("missing uptime.seconds")
	}

	secs, ok := facts["uptime.seconds"].(int)
	if !ok {
		t.Errorf("uptime.seconds should be int, got %T", facts["uptime.seconds"])
	} else if secs < 0 {
		t.Errorf("uptime.seconds should be >= 0, got %d", secs)
	}
}

func TestDisksCollector(t *testing.T) {
	c := &DisksCollector{}
	facts, err := c.Collect()
	if err != nil {
		t.Fatalf("DisksCollector error: %v", err)
	}

	// Should have at least one disk key.
	hasDisk := false
	for k := range facts {
		if strings.HasPrefix(k, "disks.") {
			hasDisk = true
			break
		}
	}
	if !hasDisk {
		t.Log("warning: no disk facts found (may be expected in some environments)")
	}
}

func TestMountpointsCollector(t *testing.T) {
	c := &MountpointsCollector{}
	facts, err := c.Collect()
	if err != nil {
		t.Fatalf("MountpointsCollector error: %v", err)
	}

	// Should have at least the root mountpoint.
	hasRoot := false
	for k := range facts {
		if strings.HasPrefix(k, "mountpoints._root.") {
			hasRoot = true
			break
		}
	}
	if !hasRoot {
		t.Error("expected root mountpoint facts")
	}
}

func TestVirtualCollector(t *testing.T) {
	c := &VirtualCollector{}
	facts, err := c.Collect()
	if err != nil {
		t.Fatalf("VirtualCollector error: %v", err)
	}

	if _, ok := facts["virtual.is_virtual"]; !ok {
		t.Error("missing virtual.is_virtual")
	}
}

func TestSSHCollector(t *testing.T) {
	c := &SSHCollector{}
	facts, err := c.Collect()
	if err != nil {
		t.Fatalf("SSHCollector error: %v", err)
	}
	// SSH keys may or may not exist, just verify no error and it's a map.
	t.Logf("SSH collector returned %d facts", len(facts))
}

func TestCloudCollector(t *testing.T) {
	c := &CloudCollector{}
	facts, err := c.Collect()
	if err != nil {
		t.Fatalf("CloudCollector error: %v", err)
	}
	// On a local machine, cloud provider is likely empty.
	t.Logf("Cloud collector returned %d facts", len(facts))
}

func TestAllCollectorsReturnValidKeys(t *testing.T) {
	collectors := All()
	for _, c := range collectors {
		facts, err := c.Collect()
		if err != nil {
			t.Errorf("collector %T returned error: %v", c, err)
			continue
		}
		for k, v := range facts {
			if k == "" {
				t.Errorf("collector %T returned empty key", c)
			}
			if !strings.Contains(k, ".") {
				t.Errorf("collector %T returned key without dot: %q", c, k)
			}
			// Values should not be nil (nil means the collector shouldn't have set it).
			if v == nil {
				t.Errorf("collector %T returned nil value for key %q", c, k)
			}
		}
	}
}
