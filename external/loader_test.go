package external

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoadTextFile(t *testing.T) {
	dir := t.TempDir()
	content := `# This is a comment
key1=value1
key2 = value2

  key3 = value3
# another comment
no_equals_line
key4=value with spaces
`
	if err := os.WriteFile(filepath.Join(dir, "facts.txt"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	facts, err := Load(dir, LoadOptions{})
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	tests := map[string]string{
		"key1": "value1",
		"key2": "value2",
		"key3": "value3",
		"key4": "value with spaces",
	}
	for k, want := range tests {
		got, ok := facts[k]
		if !ok {
			t.Errorf("missing key %q", k)
			continue
		}
		if got != want {
			t.Errorf("key %q: want %q, got %q", k, want, got)
		}
	}

	// "no_equals_line" should not appear as a key.
	if _, ok := facts["no_equals_line"]; ok {
		t.Error("line without = should be skipped")
	}
}

func TestLoadJSONFile(t *testing.T) {
	dir := t.TempDir()
	content := `{
  "app": {
    "name": "test",
    "version": "1.0",
    "nested": {
      "deep": "value"
    }
  },
  "simple": "flat"
}`
	if err := os.WriteFile(filepath.Join(dir, "facts.json"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	facts, err := Load(dir, LoadOptions{})
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	expected := map[string]any{
		"app.name":        "test",
		"app.version":     "1.0",
		"app.nested.deep": "value",
		"simple":          "flat",
	}
	for k, want := range expected {
		got, ok := facts[k]
		if !ok {
			t.Errorf("missing key %q", k)
			continue
		}
		if got != want {
			t.Errorf("key %q: want %v, got %v", k, want, got)
		}
	}
}

func TestLoadExecFact(t *testing.T) {
	dir := t.TempDir()
	script := filepath.Join(dir, "my_fact")
	content := "#!/bin/sh\necho 'exec.key1=val1'\necho 'exec.key2=val2'\n"
	if err := os.WriteFile(script, []byte(content), 0755); err != nil {
		t.Fatal(err)
	}

	facts, err := Load(dir, LoadOptions{ExecTimeout: 5 * time.Second})
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	if facts["exec.key1"] != "val1" {
		t.Errorf("expected val1, got %v", facts["exec.key1"])
	}
	if facts["exec.key2"] != "val2" {
		t.Errorf("expected val2, got %v", facts["exec.key2"])
	}
}

func TestLoadExecJSON(t *testing.T) {
	dir := t.TempDir()
	script := filepath.Join(dir, "json_fact")
	content := `#!/bin/sh
echo '{"exec_json": {"a": 1, "b": "two"}}'
`
	if err := os.WriteFile(script, []byte(content), 0755); err != nil {
		t.Fatal(err)
	}

	facts, err := Load(dir, LoadOptions{ExecTimeout: 5 * time.Second})
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	if facts["exec_json.a"] != float64(1) {
		t.Errorf("expected 1, got %v (type %T)", facts["exec_json.a"], facts["exec_json.a"])
	}
	if facts["exec_json.b"] != "two" {
		t.Errorf("expected %q, got %v", "two", facts["exec_json.b"])
	}
}

func TestLoadExecNonZeroExit(t *testing.T) {
	dir := t.TempDir()
	script := filepath.Join(dir, "bad_fact")
	content := "#!/bin/sh\nexit 1\n"
	if err := os.WriteFile(script, []byte(content), 0755); err != nil {
		t.Fatal(err)
	}

	facts, err := Load(dir, LoadOptions{ExecTimeout: 5 * time.Second})
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	// Non-zero exit should be skipped, not cause an error.
	if len(facts) != 0 {
		t.Errorf("expected empty map for non-zero exit, got %v", facts)
	}
}

func TestLoadExecTimeout(t *testing.T) {
	dir := t.TempDir()
	script := filepath.Join(dir, "slow_fact")
	// Use a trap-based script that busy-loops so the signal is delivered properly.
	content := `#!/bin/sh
trap 'exit 1' TERM INT
while true; do :; done
`
	if err := os.WriteFile(script, []byte(content), 0755); err != nil {
		t.Fatal(err)
	}

	start := time.Now()
	facts, err := Load(dir, LoadOptions{ExecTimeout: 1 * time.Second})
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	// Should have timed out within a reasonable window.
	if elapsed > 10*time.Second {
		t.Errorf("expected timeout within 10s, took %v", elapsed)
	}
	if _, ok := facts["slow.key"]; ok {
		t.Error("timed-out executable should not produce facts")
	}
}

func TestLoadNoExecOption(t *testing.T) {
	dir := t.TempDir()

	// Write a text file (should be loaded).
	if err := os.WriteFile(filepath.Join(dir, "facts.txt"), []byte("text.key=textval\n"), 0644); err != nil {
		t.Fatal(err)
	}

	// Write an executable (should be skipped).
	script := filepath.Join(dir, "exec_fact")
	if err := os.WriteFile(script, []byte("#!/bin/sh\necho 'exec.key=execval'\n"), 0755); err != nil {
		t.Fatal(err)
	}

	facts, err := Load(dir, LoadOptions{NoExec: true, ExecTimeout: 5 * time.Second})
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	if facts["text.key"] != "textval" {
		t.Errorf("text fact should be loaded, got %v", facts["text.key"])
	}
	if _, ok := facts["exec.key"]; ok {
		t.Error("executable fact should be skipped with NoExec")
	}
}

func TestLoadMissingDir(t *testing.T) {
	facts, err := Load("/nonexistent/path/that/does/not/exist", LoadOptions{})
	if err != nil {
		t.Fatalf("missing dir should not return error, got: %v", err)
	}
	if len(facts) != 0 {
		t.Errorf("expected empty map, got %v", facts)
	}
}

func TestFlatten(t *testing.T) {
	input := map[string]any{
		"top": "simple",
		"nested": map[string]any{
			"a": "va",
			"b": map[string]any{
				"c": "vc",
			},
		},
		"arr": []any{"x", "y"},
		"obj_arr": []any{
			map[string]any{"name": "first"},
			map[string]any{"name": "second"},
		},
	}

	result := make(map[string]any)
	flatten("", input, result)

	expected := map[string]any{
		"top":            "simple",
		"nested.a":       "va",
		"nested.b.c":     "vc",
		"arr.0":          "x",
		"arr.1":          "y",
		"obj_arr.0.name": "first",
		"obj_arr.1.name": "second",
	}

	for k, want := range expected {
		got, ok := result[k]
		if !ok {
			t.Errorf("missing key %q", k)
			continue
		}
		if got != want {
			t.Errorf("key %q: want %v, got %v", k, want, got)
		}
	}

	if len(result) != len(expected) {
		t.Errorf("expected %d keys, got %d: %v", len(expected), len(result), result)
	}
}
