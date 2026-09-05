package filetools

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReplaceExistingFileFailurePreservesDestination(t *testing.T) {
	dir := t.TempDir()
	destination := filepath.Join(dir, "destination.txt")
	source := filepath.Join(dir, "missing-source.txt")
	if err := os.WriteFile(destination, []byte("old"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := replaceExistingFile(source, destination); err == nil {
		t.Fatal("replaceExistingFile() unexpectedly succeeded for a missing source")
	}
	content, err := os.ReadFile(destination)
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != "old" {
		t.Fatalf("destination changed after failed replacement: %q", content)
	}
}

func TestReplaceExistingFileReplacesContent(t *testing.T) {
	dir := t.TempDir()
	destination := filepath.Join(dir, "destination.txt")
	source := filepath.Join(dir, "source.txt")
	if err := os.WriteFile(destination, []byte("old"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(source, []byte("new"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := replaceExistingFile(source, destination); err != nil {
		t.Fatalf("replaceExistingFile() error = %v", err)
	}
	content, err := os.ReadFile(destination)
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != "new" {
		t.Fatalf("destination = %q, want %q", content, "new")
	}
}
