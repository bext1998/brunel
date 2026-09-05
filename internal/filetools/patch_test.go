package filetools

import (
	"os"
	"path/filepath"
	"testing"
)

func writeFixture(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestApplyPatchReplacesLineAndReturnsUsableHash(t *testing.T) {
	dir := t.TempDir()
	path := writeFixture(t, dir, "a.txt", "one\ntwo\nthree\n")
	r := fakeResolver{root: dir}
	read, err := ReadFile(r, "a.txt", 0, 0)
	if err != nil {
		t.Fatal(err)
	}

	newHash, err := ApplyPatch(r, "a.txt", read.Hash, []Hunk{
		{StartLine: 2, EndLine: 2, OldLines: []string{"two"}, NewLines: []string{"TWO"}},
	})
	if err != nil {
		t.Fatalf("ApplyPatch() error = %v", err)
	}
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != "one\nTWO\nthree\n" {
		t.Fatalf("content = %q, want %q", content, "one\nTWO\nthree\n")
	}
	if newHash != hashBytes([]byte(content)) {
		t.Fatalf("newHash = %q, does not match written content", newHash)
	}
}

func TestApplyPatchInsertsWithoutRemovingLines(t *testing.T) {
	dir := t.TempDir()
	path := writeFixture(t, dir, "a.txt", "one\ntwo\n")
	r := fakeResolver{root: dir}
	read, err := ReadFile(r, "a.txt", 0, 0)
	if err != nil {
		t.Fatal(err)
	}

	// EndLine == StartLine-1 marks a pure insertion before line 2.
	if _, err := ApplyPatch(r, "a.txt", read.Hash, []Hunk{
		{StartLine: 2, EndLine: 1, OldLines: nil, NewLines: []string{"inserted"}},
	}); err != nil {
		t.Fatalf("ApplyPatch() error = %v", err)
	}
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != "one\ninserted\ntwo\n" {
		t.Fatalf("content = %q, want %q", content, "one\ninserted\ntwo\n")
	}
}

func TestApplyPatchDeletesLineWithEmptyNewLines(t *testing.T) {
	dir := t.TempDir()
	path := writeFixture(t, dir, "a.txt", "one\ntwo\nthree\n")
	r := fakeResolver{root: dir}
	read, err := ReadFile(r, "a.txt", 0, 0)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := ApplyPatch(r, "a.txt", read.Hash, []Hunk{
		{StartLine: 2, EndLine: 2, OldLines: []string{"two"}, NewLines: nil},
	}); err != nil {
		t.Fatalf("ApplyPatch() error = %v", err)
	}
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != "one\nthree\n" {
		t.Fatalf("content = %q, want %q", content, "one\nthree\n")
	}
}

func TestApplyPatchAppliesMultipleNonOverlappingHunks(t *testing.T) {
	dir := t.TempDir()
	path := writeFixture(t, dir, "a.txt", "one\ntwo\nthree\nfour\n")
	r := fakeResolver{root: dir}
	read, err := ReadFile(r, "a.txt", 0, 0)
	if err != nil {
		t.Fatal(err)
	}

	// Hunks are given out of order on purpose - ApplyPatch sorts by
	// StartLine before applying.
	if _, err := ApplyPatch(r, "a.txt", read.Hash, []Hunk{
		{StartLine: 4, EndLine: 4, OldLines: []string{"four"}, NewLines: []string{"FOUR"}},
		{StartLine: 1, EndLine: 1, OldLines: []string{"one"}, NewLines: []string{"ONE"}},
	}); err != nil {
		t.Fatalf("ApplyPatch() error = %v", err)
	}
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != "ONE\ntwo\nthree\nFOUR\n" {
		t.Fatalf("content = %q, want %q", content, "ONE\ntwo\nthree\nFOUR\n")
	}
}

func TestApplyPatchRejectsStaleHashAndPreservesFile(t *testing.T) {
	dir := t.TempDir()
	path := writeFixture(t, dir, "a.txt", "one\ntwo\n")
	r := fakeResolver{root: dir}
	read, err := ReadFile(r, "a.txt", 0, 0)
	if err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(path, []byte("changed\nexternally\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	_, err = ApplyPatch(r, "a.txt", read.Hash, []Hunk{
		{StartLine: 1, EndLine: 1, OldLines: []string{"one"}, NewLines: []string{"ONE"}},
	})
	if ErrorCode(err) != ErrStaleHash.Code {
		t.Fatalf("ErrorCode() = %q, want %q (err=%v)", ErrorCode(err), ErrStaleHash.Code, err)
	}
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != "changed\nexternally\n" {
		t.Fatalf("file changed after rejected stale patch: %q", content)
	}
}

func TestApplyPatchRejectsContextMismatchAndPreservesFile(t *testing.T) {
	dir := t.TempDir()
	path := writeFixture(t, dir, "a.txt", "one\ntwo\nthree\n")
	r := fakeResolver{root: dir}
	read, err := ReadFile(r, "a.txt", 0, 0)
	if err != nil {
		t.Fatal(err)
	}

	_, err = ApplyPatch(r, "a.txt", read.Hash, []Hunk{
		{StartLine: 2, EndLine: 2, OldLines: []string{"TWO (wrong context)"}, NewLines: []string{"nope"}},
	})
	if ErrorCode(err) != ErrPatchConflict.Code {
		t.Fatalf("ErrorCode() = %q, want %q (err=%v)", ErrorCode(err), ErrPatchConflict.Code, err)
	}
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != "one\ntwo\nthree\n" {
		t.Fatalf("file changed after rejected patch conflict: %q", content)
	}
}

func TestApplyPatchRejectsOverlappingHunks(t *testing.T) {
	dir := t.TempDir()
	writeFixture(t, dir, "a.txt", "one\ntwo\nthree\n")
	r := fakeResolver{root: dir}
	read, err := ReadFile(r, "a.txt", 0, 0)
	if err != nil {
		t.Fatal(err)
	}

	_, err = ApplyPatch(r, "a.txt", read.Hash, []Hunk{
		{StartLine: 1, EndLine: 2, OldLines: []string{"one", "two"}, NewLines: []string{"x"}},
		{StartLine: 2, EndLine: 2, OldLines: []string{"two"}, NewLines: []string{"y"}},
	})
	if ErrorCode(err) != ErrInvalidArgument.Code {
		t.Fatalf("ErrorCode() = %q, want %q (err=%v)", ErrorCode(err), ErrInvalidArgument.Code, err)
	}
}

func TestApplyPatchRejectsOldLinesLengthMismatch(t *testing.T) {
	dir := t.TempDir()
	writeFixture(t, dir, "a.txt", "one\ntwo\nthree\n")
	r := fakeResolver{root: dir}
	read, err := ReadFile(r, "a.txt", 0, 0)
	if err != nil {
		t.Fatal(err)
	}

	_, err = ApplyPatch(r, "a.txt", read.Hash, []Hunk{
		{StartLine: 1, EndLine: 2, OldLines: []string{"one"}, NewLines: []string{"x"}},
	})
	if ErrorCode(err) != ErrInvalidArgument.Code {
		t.Fatalf("ErrorCode() = %q, want %q (err=%v)", ErrorCode(err), ErrInvalidArgument.Code, err)
	}
}

func TestApplyPatchRejectsOutOfRangeHunk(t *testing.T) {
	dir := t.TempDir()
	writeFixture(t, dir, "a.txt", "one\ntwo\n")
	r := fakeResolver{root: dir}
	read, err := ReadFile(r, "a.txt", 0, 0)
	if err != nil {
		t.Fatal(err)
	}

	_, err = ApplyPatch(r, "a.txt", read.Hash, []Hunk{
		{StartLine: 5, EndLine: 5, OldLines: []string{"nope"}, NewLines: []string{"x"}},
	})
	if ErrorCode(err) != ErrInvalidArgument.Code {
		t.Fatalf("ErrorCode() = %q, want %q (err=%v)", ErrorCode(err), ErrInvalidArgument.Code, err)
	}
}

func TestApplyPatchPreservesNoTrailingNewlineOnFinalLineReplace(t *testing.T) {
	dir := t.TempDir()
	path := writeFixture(t, dir, "a.txt", "one\ntwo") // no trailing newline
	r := fakeResolver{root: dir}
	read, err := ReadFile(r, "a.txt", 0, 0)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := ApplyPatch(r, "a.txt", read.Hash, []Hunk{
		{StartLine: 2, EndLine: 2, OldLines: []string{"two"}, NewLines: []string{"TWO"}},
	}); err != nil {
		t.Fatalf("ApplyPatch() error = %v", err)
	}
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != "one\nTWO" {
		t.Fatalf("content = %q, want %q (no trailing newline preserved)", content, "one\nTWO")
	}
}

func TestApplyPatchRejectsMissingHunks(t *testing.T) {
	dir := t.TempDir()
	writeFixture(t, dir, "a.txt", "one\n")
	r := fakeResolver{root: dir}
	read, err := ReadFile(r, "a.txt", 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := ApplyPatch(r, "a.txt", read.Hash, nil); ErrorCode(err) != ErrInvalidArgument.Code {
		t.Fatalf("ErrorCode() = %q, want %q (err=%v)", ErrorCode(err), ErrInvalidArgument.Code, err)
	}
}
