// Package tools implements Brunel's eight frozen Alpha 1 tools.
package tools

import (
	"errors"
	"fmt"

	brunelexec "github.com/bext1998/brunel/internal/exec"
	"github.com/bext1998/brunel/internal/filetools"
	"github.com/bext1998/brunel/internal/safety"
	"github.com/bext1998/brunel/internal/workspace"
)

// Error is a stable, machine-readable tools error.
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
	ErrInvalidArgument          = &Error{Code: "E_INVALID_ARGUMENT"}
	ErrToolIO                   = &Error{Code: "E_TOOL_IO"}
	ErrWorkspaceDiffUnavailable = &Error{Code: "E_WORKSPACE_DIFF_UNAVAILABLE"}
)

func codeError(code, message string, cause error) error {
	return &Error{Code: code, Message: message, Cause: cause}
}

// ErrorCode returns a stable error code from tools errors and from dependency
// errors passed through unchanged by Registry.Call.
func ErrorCode(err error) string {
	var coded *Error
	if errors.As(err, &coded) {
		return coded.Code
	}
	for _, code := range []string{
		safety.ErrorCode(err),
		workspace.ErrorCode(err),
		filetools.ErrorCode(err),
		brunelexec.ErrorCode(err),
	} {
		if code != "" {
			return code
		}
	}
	return ""
}
