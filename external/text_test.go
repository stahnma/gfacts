package external

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadTextBasic(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.txt")
	if err := os.WriteFile(path, []byte("foo=bar\nbaz=qux\n"), 0644); err != nil {
		t.Fatal(err)
	}

	facts, err := loadText(path)
	if err != nil {
		t.Fatalf("loadText error: %v", err)
	}
	if facts["foo"] != "bar" {
		t.Errorf("expected bar, got %v", facts["foo"])
	}
	if facts["baz"] != "qux" {
		t.Errorf("expected qux, got %v", facts["baz"])
	}
}

func TestLoadTextSkipsComments(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.txt")
	content := "# comment\nkey=val\n# another\n"
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	facts, err := loadText(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(facts) != 1 {
		t.Errorf("expected 1 fact, got %d: %v", len(facts), facts)
	}
	if facts["key"] != "val" {
		t.Errorf("expected val, got %v", facts["key"])
	}
}

func TestLoadTextSkipsBlankLines(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.txt")
	content := "\n\nkey=val\n\n"
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	facts, err := loadText(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(facts) != 1 {
		t.Errorf("expected 1 fact, got %d", len(facts))
	}
}

func TestLoadTextTrimsWhitespace(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.txt")
	content := "  key  =  val  \n"
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	facts, err := loadText(path)
	if err != nil {
		t.Fatal(err)
	}
	if facts["key"] != "val" {
		t.Errorf("expected %q, got %q", "val", facts["key"])
	}
}

func TestLoadTextSkipsLinesWithoutEquals(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.txt")
	content := "no_equals\nkey=val\njust_text\n"
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	facts, err := loadText(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(facts) != 1 {
		t.Errorf("expected 1 fact, got %d: %v", len(facts), facts)
	}
}

func TestLoadTextEqualsInValue(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.txt")
	content := "key=val=ue=extra\n"
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	facts, err := loadText(path)
	if err != nil {
		t.Fatal(err)
	}
	// strings.Cut splits at first =, so value should be "val=ue=extra"
	if facts["key"] != "val=ue=extra" {
		t.Errorf("expected %q, got %q", "val=ue=extra", facts["key"])
	}
}

func TestLoadTextEmptyFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "empty.txt")
	if err := os.WriteFile(path, []byte(""), 0644); err != nil {
		t.Fatal(err)
	}

	facts, err := loadText(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(facts) != 0 {
		t.Errorf("expected 0 facts, got %d", len(facts))
	}
}
