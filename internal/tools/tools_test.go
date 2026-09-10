package tools

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

type fakeResolver struct {
	root string
}

func (f fakeResolver) Resolve(path string) (string, error) {
	if filepath.IsAbs(path) {
		return "", codeError(ErrInvalidArgument.Code, "path must be relative", nil)
	}
	return filepath.Join(f.root, filepath.Clean(path)), nil
}

func TestDecodeParamsRejectsUnknownFieldsAndWrongTypes(t *testing.T) {
	var p CreateFileParams
	err := decodeParams([]byte(`{"path":"note.txt","content":"x","extra":true}`), &p, "path", "content")
	if ErrorCode(err) != ErrInvalidArgument.Code {
		t.Fatalf("unknown field error code = %q, want %q (err=%v)", ErrorCode(err), ErrInvalidArgument.Code, err)
	}
	err = decodeParams([]byte(`{"path":"note.txt","content":7}`), &p, "path", "content")
	if ErrorCode(err) != ErrInvalidArgument.Code {
		t.Fatalf("wrong type error code = %q, want %q (err=%v)", ErrorCode(err), ErrInvalidArgument.Code, err)
	}
	err = decodeParams([]byte(`{"path":"note.txt"}`), &p, "path", "content")
	if ErrorCode(err) != ErrInvalidArgument.Code {
		t.Fatalf("missing field error code = %q, want %q (err=%v)", ErrorCode(err), ErrInvalidArgument.Code, err)
	}
}

func TestListFilesAndSearchTextUseDeterministicRelativePaths(t *testing.T) {
	root := t.TempDir()
	writeTestFile(t, filepath.Join(root, "a.txt"), "needle first\nplain\n")
	writeTestFile(t, filepath.Join(root, "nested", "b.txt"), "needle second\nneedle third\n")
	writeTestFile(t, filepath.Join(root, ".git", "ignored.txt"), "needle ignored\n")
	writeTestFile(t, filepath.Join(root, "binary.bin"), "first\x00needle\n")

	depth := 1
	entries, err := listFiles(fakeResolver{root: root}, ".", stringPtr("*.txt"), &depth)
	if err != nil {
		t.Fatalf("listFiles returned error: %v", err)
	}
	if len(entries) != 1 || entries[0].Path != "a.txt" {
		t.Fatalf("listFiles entries = %#v, want only a.txt", entries)
	}

	limit := 2
	matches, err := searchText(fakeResolver{root: root}, "needle", ".", stringPtr("*.txt"), &limit)
	if err != nil {
		t.Fatalf("searchText returned error: %v", err)
	}
	if len(matches) != 2 {
		t.Fatalf("searchText match count = %d, want 2: %#v", len(matches), matches)
	}
	if matches[0].Path != "a.txt" || matches[0].Line != 1 || matches[1].Path != "nested/b.txt" || matches[1].Line != 1 {
		t.Fatalf("searchText matches = %#v, want deterministic path/line order", matches)
	}

	_, err = searchText(fakeResolver{root: root}, "[", ".", nil, nil)
	if !errors.Is(err, ErrInvalidArgument) {
		t.Fatalf("invalid regexp error = %v, want %v", err, ErrInvalidArgument)
	}
}

func TestRequiredHashValidationHasNoFileSideEffect(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "note.txt")
	const original = "original\n"
	writeTestFile(t, path, original)

	if err := validateWriteFileParams(WriteFileParams{Path: "note.txt", Content: "changed"}); !errors.Is(err, ErrInvalidArgument) {
		t.Fatalf("write validation error = %v, want %v", err, ErrInvalidArgument)
	}
	if err := validateApplyPatchParams(ApplyPatchParams{Path: "note.txt"}); !errors.Is(err, ErrInvalidArgument) {
		t.Fatalf("patch validation error = %v, want %v", err, ErrInvalidArgument)
	}
	actual, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read unchanged file: %v", err)
	}
	if string(actual) != original {
		t.Fatalf("file changed during validation: got %q, want %q", actual, original)
	}
}

func writeTestFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("create test directory: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write test file: %v", err)
	}
}

func stringPtr(value string) *string { return &value }
