package external

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadJSONBasic(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.json")
	content := `{"key1": "val1", "key2": 42}`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	facts, err := loadJSON(path)
	if err != nil {
		t.Fatalf("loadJSON error: %v", err)
	}
	if facts["key1"] != "val1" {
		t.Errorf("expected val1, got %v", facts["key1"])
	}
	if facts["key2"] != float64(42) {
		t.Errorf("expected 42, got %v (type %T)", facts["key2"], facts["key2"])
	}
}

func TestLoadJSONNested(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.json")
	content := `{"a": {"b": {"c": "deep"}}}`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	facts, err := loadJSON(path)
	if err != nil {
		t.Fatal(err)
	}
	if facts["a.b.c"] != "deep" {
		t.Errorf("expected deep, got %v", facts["a.b.c"])
	}
}

func TestLoadJSONArrays(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.json")
	content := `{"items": ["a", "b", "c"]}`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	facts, err := loadJSON(path)
	if err != nil {
		t.Fatal(err)
	}
	if facts["items.0"] != "a" {
		t.Errorf("expected a, got %v", facts["items.0"])
	}
	if facts["items.1"] != "b" {
		t.Errorf("expected b, got %v", facts["items.1"])
	}
	if facts["items.2"] != "c" {
		t.Errorf("expected c, got %v", facts["items.2"])
	}
}

func TestLoadJSONArrayOfObjects(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.json")
	content := `{"users": [{"name": "alice"}, {"name": "bob"}]}`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	facts, err := loadJSON(path)
	if err != nil {
		t.Fatal(err)
	}
	if facts["users.0.name"] != "alice" {
		t.Errorf("expected alice, got %v", facts["users.0.name"])
	}
	if facts["users.1.name"] != "bob" {
		t.Errorf("expected bob, got %v", facts["users.1.name"])
	}
}

func TestLoadJSONInvalidJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.json")
	if err := os.WriteFile(path, []byte("not json {{{"), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := loadJSON(path)
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestLoadJSONBooleans(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.json")
	content := `{"flag": true, "other": false, "nothing": null}`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	facts, err := loadJSON(path)
	if err != nil {
		t.Fatal(err)
	}
	if facts["flag"] != true {
		t.Errorf("expected true, got %v", facts["flag"])
	}
	if facts["other"] != false {
		t.Errorf("expected false, got %v", facts["other"])
	}
	if facts["nothing"] != nil {
		t.Errorf("expected nil, got %v", facts["nothing"])
	}
}

func TestFlattenEmptyMap(t *testing.T) {
	result := make(map[string]any)
	flatten("", map[string]any{}, result)
	if len(result) != 0 {
		t.Errorf("expected empty result, got %v", result)
	}
}

func TestFlattenWithPrefix(t *testing.T) {
	result := make(map[string]any)
	flatten("root", map[string]any{
		"a": "va",
		"b": map[string]any{"c": "vc"},
	}, result)

	if result["root.a"] != "va" {
		t.Errorf("expected va, got %v", result["root.a"])
	}
	if result["root.b.c"] != "vc" {
		t.Errorf("expected vc, got %v", result["root.b.c"])
	}
}
