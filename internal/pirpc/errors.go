// Package pirpc implements the passthrough and translation responsibilities
// Issue #8 (F-7, ADR-002) assigns to Brunel for the Pi RPC subprocess:
// building the `pi --mode rpc` launch arguments from the user's chosen
// provider/model, injecting credentials via environment variables (never a
// project file), and translating Pi-reported provider-layer errors into
// Brunel's own stable error codes. Brunel does not re-implement SSE
// parsing, tool-call probing, or retry/backoff - that is delegated to Pi.
//
// Starting and managing the Pi RPC subprocess itself, and translating its
// RPC *events* (message/tool-call/usage) into agent.Event, is Issue #9's
// responsibility; this package only provides the pieces #9 builds on. Per
// INV-9 (spec.md §10, `[FROZEN]`), no code in this package may ever send a
// bash-type RPC command (the JSON `type` field literally set to `bash`) -
// see bash_guard_test.go.
package pirpc

import (
	"errors"
	"fmt"
)

// Error is a stable, machine-readable pirpc error.
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

	// Provider-layer error codes a Pi RPC error is translated to
	// (spec.md §5.3/§9 CT-6): authentication, quota/rate-limit, an
	// unknown/unsupported model, and a malformed or unexpected RPC
	// protocol response. ErrPiProviderError is the fallback for a
	// provider-layer failure that does not match any of those.
	ErrPiProviderAuth  = &Error{Code: "E_PI_PROVIDER_AUTH"}
	ErrPiProviderQuota = &Error{Code: "E_PI_PROVIDER_QUOTA"}
	ErrPiModelNotFound = &Error{Code: "E_PI_MODEL_NOT_FOUND"}
	ErrPiProtocol      = &Error{Code: "E_PI_PROTOCOL"}
	ErrPiProviderError = &Error{Code: "E_PI_PROVIDER_ERROR"}
)

func codeError(code, message string, cause error) error {
	return &Error{Code: code, Message: message, Cause: cause}
}

// ErrorCode returns a stable error code when err is a pirpc Error.
func ErrorCode(err error) string {
	var coded *Error
	if errors.As(err, &coded) {
		return coded.Code
	}
	return ""
}
