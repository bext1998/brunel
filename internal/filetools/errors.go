// Package filetools implements the stale-read hash guard and atomic write
// path shared by the read_file, create_file, write_file and apply_patch
// tools (Alpha 1 Issue #5 / TASK-A1-F04). It intentionally does not resolve
// workspace paths itself: callers pass a Resolver (satisfied structurally by
// *workspace.Workspace) so the hash/patch/atomic-write logic can be tested
// on any platform even though workspace path resolution is Windows-only.
package filetools

import (
	"errors"
	"fmt"
)

// Error is a stable, machine-readable filetools error.
type Error struct {
	Code    string
	Message string
	Cause   error
}

func (e *Error) Error() string {
	if e == nil {
		return ""
	}
	if e.Message == "" {
		return e.Code
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e *Error) Unwrap() error { return e.Cause }

func (e *Error) Is(target error) bool {
	t, ok := target.(*Error)
	return ok && e != nil && t != nil && e.Code == t.Code
}

var (
	ErrInvalidArgument = &Error{Code: "E_INVALID_ARGUMENT"}
	ErrNotFound        = &Error{Code: "E_FILE_NOT_FOUND"}
	ErrFileExists      = &Error{Code: "E_FILE_EXISTS"}
	ErrStaleHash       = &Error{Code: "E_STALE_HASH"}
	ErrPatchConflict   = &Error{Code: "E_PATCH_CONFLICT"}
	ErrFileIO          = &Error{Code: "E_FILE_IO"}
)

func codeError(code, message string, cause error) error {
	return &Error{Code: code, Message: message, Cause: cause}
}

// ErrorCode returns a stable error code when err is a filetools Error. A
// path Resolver's own error (e.g. workspace's E_PATH_ESCAPE) is returned
// unwrapped by every function in this package, so callers that need its
// code should use the resolver's own ErrorCode helper (errors.As also
// works, since resolver errors are never wrapped by filetools.Error).
func ErrorCode(err error) string {
	var coded *Error
	if errors.As(err, &coded) {
		return coded.Code
	}
	return ""
}
