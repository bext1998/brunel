package exec

import (
	"errors"
	"fmt"
)

// Error is a stable, machine-readable exec error.
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
	ErrInvalidArgument     = &Error{Code: "E_INVALID_ARGUMENT"}
	ErrToolTimeout         = &Error{Code: "E_TOOL_TIMEOUT"}
	ErrPwshRequired        = &Error{Code: "E_PWSH_REQUIRED"}
	ErrUnsupportedPlatform = &Error{Code: "E_UNSUPPORTED_PLATFORM"}
	ErrExecInternal        = &Error{Code: "E_EXEC_INTERNAL"}
)

func codeError(code, message string, cause error) error {
	return &Error{Code: code, Message: message, Cause: cause}
}

// ErrorCode returns a stable error code when err is an exec Error.
func ErrorCode(err error) string {
	var coded *Error
	if errors.As(err, &coded) {
		return coded.Code
	}
	return ""
}
