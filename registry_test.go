package gfacts

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestProgrammaticFactPrecedence(t *testing.T) {
	// Create a temp dir with an external fact that sets "precedence.test".
	dir := t.TempDir()
	err := os.WriteFile(filepath.Join(dir, "test.txt"), []byte("precedence.test=external_value\n"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Also register a programmatic fact with the same key.
	Register("precedence.test", "programmatic_value")
	defer func() {
		globalRegistry.mu.Lock()
		delete(globalRegistry.static, "precedence.test")
		globalRegistry.mu.Unlock()
	}()

	facts, err := GatherWithOptions(Options{
		ExternalDirs: []string{dir},
	})
	if err != nil {
		t.Fatalf("Gather returned error: %v", err)
	}

	// Programmatic should win over external.
	if facts["precedence.test"] != "programmatic_value" {
		t.Errorf("expected programmatic to win, got %v", facts["precedence.test"])
	}
}

func TestExternalOverridesCore(t *testing.T) {
	// Create an external fact that overrides os.name.
	dir := t.TempDir()
	err := os.WriteFile(filepath.Join(dir, "override.txt"), []byte("os.name=ExternalOS\n"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Ensure no programmatic override.
	globalRegistry.mu.Lock()
	delete(globalRegistry.static, "os.name")
	globalRegistry.mu.Unlock()

	facts, err := GatherWithOptions(Options{
		ExternalDirs: []string{dir},
	})
	if err != nil {
		t.Fatalf("Gather returned error: %v", err)
	}

	// External should override core.
	if facts["os.name"] != "ExternalOS" {
		t.Errorf("expected external to override core, got %v", facts["os.name"])
	}
}

func TestRegistryResolve(t *testing.T) {
	r := newRegistry(Options{NoExternal: true})
	facts, err := r.resolve()
	if err != nil {
		t.Fatalf("resolve returned error: %v", err)
	}
	if len(facts) == 0 {
		t.Error("resolve returned empty map")
	}

	// Verify the map has only string keys.
	for k := range facts {
		if k == "" {
			t.Error("found empty key in resolved facts")
		}
		// Keys should be dotted notation.
		if !strings.Contains(k, ".") {
			t.Logf("warning: key %q has no dot separator", k)
		}
	}
}
