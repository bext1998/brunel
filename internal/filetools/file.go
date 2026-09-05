package filetools

import (
	"os"
	"path/filepath"
	"strings"
)

// Resolver resolves a workspace-relative path to an absolute path, rejecting
// any path that escapes the bound workspace root. *workspace.Workspace
// satisfies this interface; tests substitute a fake resolver so the
// hash/patch/atomic-write logic in this package can run on any platform.
type Resolver interface {
	Resolve(path string) (string, error)
}

// Line is one line of file content, numbered from 1, with its line
// terminator stripped from Text.
type Line struct {
	Number int
	Text   string
}

// ReadResult is the outcome of ReadFile: the requested line range (or the
// whole file when no range is given) together with the SHA-256 hash of the
// complete, unmodified file. The hash always covers the whole file
// regardless of the requested range, so it can be passed back later as
// expected_hash to WriteFile or ApplyPatch.
type ReadResult struct {
	Lines []Line
	Hash  string
}

// ReadFile reads the file at path (resolved through r) and returns the
// requested 1-based, inclusive line range together with the whole-file
// SHA-256 hash. startLine and/or endLine of 0 mean "from the first" /
// "to the last" line respectively.
func ReadFile(r Resolver, path string, startLine, endLine int) (ReadResult, error) {
	if startLine < 0 || endLine < 0 {
		return ReadResult{}, codeError(ErrInvalidArgument.Code, "start_line/end_line must not be negative", nil)
	}
	if startLine > 0 && endLine > 0 && endLine < startLine {
		return ReadResult{}, codeError(ErrInvalidArgument.Code, "end_line is before start_line", nil)
	}
	abs, err := r.Resolve(path)
	if err != nil {
		return ReadResult{}, err
	}
	data, err := readExistingFile(abs)
	if err != nil {
		return ReadResult{}, err
	}
	hash := hashBytes(data)
	lines := splitLines(data)
	from, to := clampRange(len(lines), startLine, endLine)
	result := make([]Line, 0, to-from+1)
	for i := from; i <= to; i++ {
		result = append(result, Line{Number: i, Text: lines[i-1].text})
	}
	return ReadResult{Lines: result, Hash: hash}, nil
}

func clampRange(total, startLine, endLine int) (from, to int) {
	from = startLine
	if from <= 0 {
		from = 1
	}
	to = endLine
	if to <= 0 || to > total {
		to = total
	}
	if from > total {
		// Requested range starts past the end of the file: empty result,
		// not an error - the whole-file hash is still meaningful on its own.
		return 1, 0
	}
	return from, to
}

// CreateFile creates a new file at path with content. It fails, without any
// side effect, if the file already exists - create_file never overwrites
// existing content.
func CreateFile(r Resolver, path, content string) (string, error) {
	abs, err := r.Resolve(path)
	if err != nil {
		return "", err
	}
	if _, statErr := os.Lstat(abs); statErr == nil {
		return "", codeError(ErrFileExists.Code, "file already exists", nil)
	} else if !os.IsNotExist(statErr) {
		return "", codeError(ErrFileIO.Code, "cannot inspect target path", statErr)
	}
	data := []byte(content)
	file, err := os.OpenFile(abs, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
	if err != nil {
		if os.IsExist(err) {
			return "", codeError(ErrFileExists.Code, "file already exists", nil)
		}
		return "", codeError(ErrFileIO.Code, "cannot create file", err)
	}
	if _, err := file.Write(data); err != nil {
		_ = file.Close()
		_ = os.Remove(abs)
		return "", codeError(ErrFileIO.Code, "cannot write new file", err)
	}
	if err := file.Sync(); err != nil {
		_ = file.Close()
		_ = os.Remove(abs)
		return "", codeError(ErrFileIO.Code, "cannot sync new file", err)
	}
	if err := file.Close(); err != nil {
		_ = os.Remove(abs)
		return "", codeError(ErrFileIO.Code, "cannot close new file", err)
	}
	return hashBytes(data), nil
}

// WriteFile overwrites an existing file's full content, but only when its
// current whole-file hash still matches expectedHash. Any failure -
// including a stale hash - leaves the target file completely unchanged.
func WriteFile(r Resolver, path, expectedHash, content string) (string, error) {
	if strings.TrimSpace(expectedHash) == "" {
		return "", codeError(ErrInvalidArgument.Code, "expected_hash is required", nil)
	}
	abs, err := r.Resolve(path)
	if err != nil {
		return "", err
	}
	current, err := readExistingFile(abs)
	if err != nil {
		return "", err
	}
	if hashBytes(current) != expectedHash {
		return "", codeError(ErrStaleHash.Code, "file changed since it was last read", nil)
	}
	newData := []byte(content)
	if err := atomicReplace(abs, newData, expectedHash); err != nil {
		return "", err
	}
	return hashBytes(newData), nil
}

// readExistingFile reads path, mapping "does not exist" and "is a
// directory" to stable filetools error codes.
func readExistingFile(path string) ([]byte, error) {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, codeError(ErrNotFound.Code, "file does not exist", err)
		}
		return nil, codeError(ErrFileIO.Code, "cannot stat file", err)
	}
	if info.IsDir() {
		return nil, codeError(ErrInvalidArgument.Code, "path is a directory", nil)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, codeError(ErrFileIO.Code, "cannot read file", err)
	}
	return data, nil
}

// atomicReplace writes data to a temporary file next to target, re-checks
// target's hash immediately before the swap to narrow the stale-write race
// as much as a single-process, non-transactional filesystem allows, and
// only then atomically replaces target. On any failure the temporary file
// is removed and target is left byte-for-byte unchanged.
func atomicReplace(target string, data []byte, expectedHash string) error {
	dir := filepath.Dir(target)
	tmp, err := os.CreateTemp(dir, ".brunel-filetools-*")
	if err != nil {
		return codeError(ErrFileIO.Code, "cannot create temporary file", err)
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return codeError(ErrFileIO.Code, "cannot write temporary file", err)
	}
	if err := tmp.Sync(); err != nil {
		_ = tmp.Close()
		return codeError(ErrFileIO.Code, "cannot sync temporary file", err)
	}
	if err := tmp.Close(); err != nil {
		return codeError(ErrFileIO.Code, "cannot close temporary file", err)
	}
	if expectedHash != "" {
		latest, err := os.ReadFile(target)
		if err != nil {
			return codeError(ErrFileIO.Code, "cannot re-read file before replace", err)
		}
		if hashBytes(latest) != expectedHash {
			return codeError(ErrStaleHash.Code, "file changed since it was last read", nil)
		}
	}
	if err := replaceExistingFile(tmpName, target); err != nil {
		return codeError(ErrFileIO.Code, "cannot replace file", err)
	}
	return nil
}
