package gfacts

import (
	"strings"
	"testing"
)

func TestGather(t *testing.T) {
	facts, err := GatherWithOptions(Options{
		NoExternal: true,
	})
	if err != nil {
		t.Fatalf("Gather returned error: %v", err)
	}
	if len(facts) == 0 {
		t.Fatal("Gather returned empty map")
	}

	// Verify some expected key prefixes exist.
	expectedPrefixes := []string{"os.", "kernel.", "networking.", "memory."}
	for _, prefix := range expectedPrefixes {
		found := false
		for k := range facts {
			if strings.HasPrefix(k, prefix) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected at least one key with prefix %q, found none", prefix)
		}
	}
}

func TestValue(t *testing.T) {
	// Register a known fact so we can test exact lookup.
	Register("test.value.exact", "hello")
	defer func() {
		globalRegistry.mu.Lock()
		delete(globalRegistry.static, "test.value.exact")
		globalRegistry.mu.Unlock()
	}()

	val, err := Value("test.value.exact")
	if err != nil {
		t.Fatalf("Value returned error: %v", err)
	}
	if val != "hello" {
		t.Errorf("expected %q, got %v", "hello", val)
	}
}

func TestValuePrefixLookup(t *testing.T) {
	// Register facts under a common prefix.
	Register("testprefix.a", "1")
	Register("testprefix.b", "2")
	defer func() {
		globalRegistry.mu.Lock()
		delete(globalRegistry.static, "testprefix.a")
		delete(globalRegistry.static, "testprefix.b")
		globalRegistry.mu.Unlock()
	}()

	val, err := Value("testprefix")
	if err != nil {
		t.Fatalf("Value returned error: %v", err)
	}
	sub, ok := val.(map[string]any)
	if !ok {
		t.Fatalf("expected map[string]any for prefix lookup, got %T", val)
	}
	if sub["testprefix.a"] != "1" || sub["testprefix.b"] != "2" {
		t.Errorf("unexpected sub-map contents: %v", sub)
	}
}

func TestValueNotFound(t *testing.T) {
	val, err := Value("nonexistent.key.that.does.not.exist.12345")
	if err != nil {
		t.Fatalf("Value returned error: %v", err)
	}
	if val != nil {
		t.Errorf("expected nil for missing key, got %v", val)
	}
}

func TestRegister(t *testing.T) {
	Register("custom.static.fact", 42)
	defer func() {
		globalRegistry.mu.Lock()
		delete(globalRegistry.static, "custom.static.fact")
		globalRegistry.mu.Unlock()
	}()

	facts, err := GatherWithOptions(Options{NoExternal: true})
	if err != nil {
		t.Fatalf("Gather returned error: %v", err)
	}
	if facts["custom.static.fact"] != 42 {
		t.Errorf("expected 42, got %v", facts["custom.static.fact"])
	}
}

func TestRegisterFunc(t *testing.T) {
	RegisterFunc("custom.dynamic.fact", func() (any, error) {
		return "dynamic_value", nil
	})
	defer func() {
		globalRegistry.mu.Lock()
		delete(globalRegistry.funcs, "custom.dynamic.fact")
		globalRegistry.mu.Unlock()
	}()

	facts, err := GatherWithOptions(Options{NoExternal: true})
	if err != nil {
		t.Fatalf("Gather returned error: %v", err)
	}
	if facts["custom.dynamic.fact"] != "dynamic_value" {
		t.Errorf("expected %q, got %v", "dynamic_value", facts["custom.dynamic.fact"])
	}
}

func TestRegisterOverridesCore(t *testing.T) {
	// os.name is a core fact. Register a programmatic override.
	Register("os.name", "OverriddenOS")
	defer func() {
		globalRegistry.mu.Lock()
		delete(globalRegistry.static, "os.name")
		globalRegistry.mu.Unlock()
	}()

	facts, err := GatherWithOptions(Options{NoExternal: true})
	if err != nil {
		t.Fatalf("Gather returned error: %v", err)
	}
	if facts["os.name"] != "OverriddenOS" {
		t.Errorf("programmatic fact should override core; got %v", facts["os.name"])
	}
}

func TestLookup(t *testing.T) {
	facts := map[string]any{
		"a.b.c": "val1",
		"a.b.d": "val2",
		"a.x":   "val3",
		"z":     "val4",
	}

	// Exact match.
	if v := lookup(facts, "z"); v != "val4" {
		t.Errorf("exact match: expected %q, got %v", "val4", v)
	}

	// Exact match with dotted key.
	if v := lookup(facts, "a.b.c"); v != "val1" {
		t.Errorf("exact dotted match: expected %q, got %v", "val1", v)
	}

	// Prefix match.
	v := lookup(facts, "a.b")
	sub, ok := v.(map[string]any)
	if !ok {
		t.Fatalf("prefix match: expected map, got %T", v)
	}
	if len(sub) != 2 {
		t.Errorf("prefix match: expected 2 entries, got %d", len(sub))
	}

	// No match.
	if v := lookup(facts, "nonexistent"); v != nil {
		t.Errorf("no match: expected nil, got %v", v)
	}
}
