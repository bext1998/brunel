package session

import "github.com/bext1998/brunel/internal/redact"

// MaskSecrets applies best-effort masking to known credential forms. It is
// intentionally not a claim that arbitrary sensitive content can be detected.
// The implementation lives in internal/redact so the same masking is shared
// with other trust boundaries (e.g. internal/pirpc provider errors).
func MaskSecrets(value string) string {
	return redact.Secrets(value)
}
