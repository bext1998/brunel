// Package redact provides best-effort masking of known credential forms in
// strings that are about to cross a trust boundary (persisted to disk,
// surfaced in a public error, emitted as an event). It is a leaf utility
// with no internal dependencies so any layer can use it; it is intentionally
// not a claim that arbitrary sensitive content can be detected.
package redact

import (
	"regexp"
	"strings"
)

var (
	bearerSecretRE = regexp.MustCompile(`(?i)(Authorization\s*:\s*Bearer\s+)[^\s"']+`)
	jsonAuthRE     = regexp.MustCompile(`(?i)("(?:authorization|api[_-]?key|token|password|secret)"\s*:\s*")[^"]*(")`)
	apiKeyRE       = regexp.MustCompile(`(?i)\b(?:sk-or-v1|sk|or-v1)-[A-Za-z0-9_-]{8,}`)
	envSecretRE    = regexp.MustCompile(`(?mi)(^\s*[A-Za-z0-9_]*(?:API[_-]?KEY|TOKEN|SECRET|PASSWORD|AUTHORIZATION)[A-Za-z0-9_]*\s*=\s*)[^\r\n#]+`)
)

// minKnownSecretLen guards the exact-match pass in Secrets: replacing a very
// short or empty string everywhere it appears would corrupt unrelated text
// (and ReplaceAll of "" injects the marker between every rune). Real API
// keys and tokens are far longer than this floor.
const minKnownSecretLen = 8

// Secrets applies best-effort masking to known credential forms: an
// Authorization: Bearer header, a JSON authorization/api_key/token/password/
// secret value, a bare sk-/or-v1- style API key, and a KEY=VALUE line whose
// key name looks like a credential.
//
// known holds exact credential values the caller already possesses (for
// example the API key it just injected into the subprocess environment).
// Each is redacted by exact substring match before the heuristics run, so
// provider key formats the heuristics do not recognize (Google "AIza...",
// Groq "gsk_...", etc.) are still removed when their value is known. Values
// shorter than minKnownSecretLen are ignored. Unknown secret shapes with no
// matching known value pass through unchanged.
func Secrets(value string, known ...string) string {
	for _, secret := range known {
		secret = strings.TrimSpace(secret)
		if len(secret) < minKnownSecretLen {
			continue
		}
		value = strings.ReplaceAll(value, secret, "[REDACTED]")
	}
	value = bearerSecretRE.ReplaceAllString(value, `${1}[REDACTED]`)
	value = jsonAuthRE.ReplaceAllString(value, `${1}[REDACTED]${2}`)
	value = envSecretRE.ReplaceAllString(value, `${1}[REDACTED]`)
	return apiKeyRE.ReplaceAllString(value, "[REDACTED]")
}
