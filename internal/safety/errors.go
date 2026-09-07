// Package safety implements the single incident-prevention decision gate
// every tool call must pass through before any I/O (Alpha 1 Issue #7 /
// TASK-A1-F06, spec.md §5.2/§6). It is not a sandbox: PowerShell is a full
// programming language, and classification here is a best-effort string and
// token judgment over a short, representative list of command forms - it
// never claims complete PowerShell semantic coverage (spec.md §6.2).
package safety

import (
	"errors"
	"fmt"
)

// Error is a stable, machine-readable safety error.
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
	// ErrReadonlyMode is the deterministic rejection (spec.md §6.3) for any
	// mutating tool attempted while the gate is in readonly mode - a tool
	// precondition, not a risk classification, so it is never confirmable.
	ErrReadonlyMode = &Error{Code: "E_READONLY_MODE"}
	// ErrApprovalRequiredNoTTY is spec.md's own named code (§6.3, EC-6):
	// a CONFIRM decision with no Approver available ends the run instead
	// of blocking.
	ErrApprovalRequiredNoTTY = &Error{Code: "E_APPROVAL_REQUIRED_NO_TTY"}
	// ErrApprovalDenied is returned when the user explicitly declines a
	// CONFIRM prompt.
	ErrApprovalDenied = &Error{Code: "E_APPROVAL_DENIED"}
)

func codeError(code, message string, cause error) error {
	return &Error{Code: code, Message: message, Cause: cause}
}

// ErrorCode returns a stable error code when err is a safety Error.
func ErrorCode(err error) string {
	var coded *Error
	if errors.As(err, &coded) {
		return coded.Code
	}
	return ""
}
