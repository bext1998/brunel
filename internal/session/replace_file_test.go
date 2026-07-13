package session

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReplaceExistingFileFailurePreservesDestination(t *testing.T) {
	dir := t.TempDir()
	destination := filepath.Join(dir, "destination.json")
	source := filepath.Join(dir, "missing.json")
	if err := os.WriteFile(destination, []byte("old"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := replaceExistingFile(source, destination); err == nil {
		t.Fatal("replaceExistingFile() unexpectedly succeeded for missing source")
	}
	content, err := os.ReadFile(destination)
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != "old" {
		t.Fatalf("destination changed after failed replacement: %q", content)
	}
}

func TestWriteJSONAtomicReplacesExistingFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "value.json")
	if err := writeJSONAtomic(path, map[string]int{"value": 1}); err != nil {
		t.Fatal(err)
	}
	if err := writeJSONAtomic(path, map[string]int{"value": 2}); err != nil {
		t.Fatal(err)
	}
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != "{\n  \"value\": 2\n}\n" {
		t.Fatalf("unexpected replaced JSON: %q", content)
	}
}
