// Package gfacts gathers system facts (hardware, software, networking, etc.)
// and returns them as a flat map with dotted keys.
package gfacts

import "time"

// FactFunc is a function that dynamically resolves a fact value.
type FactFunc func() (any, error)

// Options controls fact resolution behavior.
type Options struct {
	// ExternalDirs lists directories to scan for external facts.
	// Default: ["/etc/gfacts/gfacts.d"]
	ExternalDirs []string

	// NoExternal skips all external fact loading.
	NoExternal bool

	// NoCustomExec loads static external files but skips executables.
	NoCustomExec bool

	// ExecTimeout is the timeout for running executable facts.
	// Default: 30s.
	ExecTimeout time.Duration
}

// DefaultOptions returns Options with sensible defaults.
func DefaultOptions() Options {
	return Options{
		ExternalDirs: []string{"/etc/gfacts/gfacts.d"},
		ExecTimeout:  30 * time.Second,
	}
}

// Gather resolves all facts and returns them as a flat map with dotted keys.
func Gather() (map[string]any, error) {
	return GatherWithOptions(DefaultOptions())
}

// GatherWithOptions resolves all facts using the provided options.
func GatherWithOptions(opts Options) (map[string]any, error) {
	r := newRegistry(opts)
	return r.resolve()
}

// Value returns a single fact by dotted key. If the key is a prefix,
// returns a map[string]any of all facts under that prefix.
func Value(key string) (any, error) {
	facts, err := Gather()
	if err != nil {
		return nil, err
	}
	return lookup(facts, key), nil
}

// Register adds a static programmatic fact that takes precedence over
// core and external facts.
func Register(key string, value any) {
	globalRegistry.registerStatic(key, value)
}

// RegisterFunc adds a dynamically resolved programmatic fact.
func RegisterFunc(key string, fn FactFunc) {
	globalRegistry.registerFunc(key, fn)
}
