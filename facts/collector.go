// Package facts implements core system fact collectors.
package facts

// Collector gathers a set of related system facts.
type Collector interface {
	Collect() (map[string]any, error)
}

// All returns all registered collectors for the current platform.
// Platform-specific collectors are added via build-tagged files.
func All() []Collector {
	return platformCollectors()
}
