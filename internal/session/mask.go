package session

import "regexp"

var (
	bearerSecretRE = regexp.MustCompile(`(?i)(Authorization\s*:\s*Bearer\s+)[^\s"']+`)
	jsonAuthRE     = regexp.MustCompile(`(?i)("(?:authorization|api[_-]?key|token|password|secret)"\s*:\s*")[^"]*(")`)
	apiKeyRE       = regexp.MustCompile(`(?i)\b(?:sk-or-v1|sk|or-v1)-[A-Za-z0-9_-]{8,}`)
	envSecretRE    = regexp.MustCompile(`(?mi)(^\s*[A-Za-z0-9_]*(?:API[_-]?KEY|TOKEN|SECRET|PASSWORD|AUTHORIZATION)[A-Za-z0-9_]*\s*=\s*)[^\r\n#]+`)
)

// MaskSecrets applies best-effort masking to known credential forms. It is
// intentionally not a claim that arbitrary sensitive content can be detected.
func MaskSecrets(value string) string {
	value = bearerSecretRE.ReplaceAllString(value, `${1}[REDACTED]`)
	value = jsonAuthRE.ReplaceAllString(value, `${1}[REDACTED]${2}`)
	value = envSecretRE.ReplaceAllString(value, `${1}[REDACTED]`)
	return apiKeyRE.ReplaceAllString(value, "[REDACTED]")
}
