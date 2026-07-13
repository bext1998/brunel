package config

import (
	"errors"
	"fmt"
)

type Error struct {
	Code   string
	Source string
	Cause  error
}

func (e *Error) Error() string {
	if e == nil {
		return ""
	}
	if e.Source == "" {
		return e.Code
	}
	return fmt.Sprintf("%s (%s)", e.Code, e.Source)
}

func (e *Error) Unwrap() error { return e.Cause }

func (e *Error) Is(target error) bool {
	t, ok := target.(*Error)
	return ok && e != nil && t != nil && e.Code == t.Code
}

var (
	ErrConfigInvalid       = &Error{Code: "E_CONFIG_INVALID"}
	ErrConfigCredential    = &Error{Code: "E_CONFIG_CREDENTIAL"}
	ErrUnsupportedPlatform = &Error{Code: "E_UNSUPPORTED_PLATFORM"}
)

func configError(code, source string, cause error) error {
	return &Error{Code: code, Source: source, Cause: cause}
}

func ErrorCode(err error) string {
	var coded *Error
	if errors.As(err, &coded) {
		return coded.Code
	}
	return ""
}
