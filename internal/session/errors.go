package session

import (
	"errors"
	"fmt"
)

// Error is a stable, machine-readable session error.
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
	ErrInvalidArgument   = &Error{Code: "E_INVALID_ARGUMENT"}
	ErrSessionNotFound   = &Error{Code: "E_SESSION_NOT_FOUND"}
	ErrSessionAmbiguous  = &Error{Code: "E_SESSION_AMBIGUOUS"}
	ErrSessionCorrupt    = &Error{Code: "E_SESSION_CORRUPT"}
	ErrEventLogTail      = &Error{Code: "E_EVENT_LOG_TAIL"}
	ErrSessionClosed     = &Error{Code: "E_SESSION_CLOSED"}
	ErrImmutableMetadata = &Error{Code: "E_SESSION_IMMUTABLE"}
)

func codeError(code, message string, cause error) error {
	return &Error{Code: code, Message: message, Cause: cause}
}

// ErrorCode returns a stable error code when err is a session Error.
func ErrorCode(err error) string {
	var coded *Error
	if errors.As(err, &coded) {
		return coded.Code
	}
	return ""
}
