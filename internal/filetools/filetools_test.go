package filetools

import "path/filepath"

// fakeResolver joins path onto root without any of workspace's real escape
// or identity checks - those are workspace's own responsibility (Issue #3)
// and are exercised by workspace's own tests. It lets this package's
// hash/patch/atomic-write logic be tested on any platform.
type fakeResolver struct {
	root string
}

func (f fakeResolver) Resolve(path string) (string, error) {
	if filepath.IsAbs(path) {
		return "", codeError(ErrInvalidArgument.Code, "path must be relative to workspace root", nil)
	}
	return filepath.Join(f.root, filepath.Clean(path)), nil
}
