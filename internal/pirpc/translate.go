package pirpc

import (
	"strings"

	"github.com/bext1998/brunel/internal/redact"
)

// ProviderErrorReport is the minimal shape of a provider-layer failure
// reported over Pi's RPC protocol that #9's event loop will decode: an
// optional machine-readable Kind (whatever short category string, if any,
// Pi's own error event carries) and the human-readable Message it reported.
// Neither field's exact wire format is specified by Pi - this package does
// not parse RPC JSON itself (that is Issue #9's responsibility) - so
// TranslateProviderError is deliberately best-effort, the same spirit as
// internal/safety's command classification (spec.md §5.3/§9 CT-6: "Brunel
// 不重新實作重試邏輯，也不得覆蓋或攔截 Pi 已決定的重試／放棄行為" - this
// package only labels a failure for display, it never retries or hides it).
type ProviderErrorReport struct {
	Kind    string
	Message string
}

// knownKinds maps normalized, Pi-reported category strings that are
// reasonably expected for an error event to Brunel's stable codes. Real
// values are confirmed against Pi's actual RPC protocol as #9 implements
// the event decoder; unrecognized or empty Kind falls back to scanning
// Message.
var knownKinds = map[string]*Error{
	"auth":               ErrPiProviderAuth,
	"authentication":     ErrPiProviderAuth,
	"unauthorized":       ErrPiProviderAuth,
	"invalid_api_key":    ErrPiProviderAuth,
	"quota":              ErrPiProviderQuota,
	"rate_limit":         ErrPiProviderQuota,
	"insufficient_quota": ErrPiProviderQuota,
	"model_not_found":    ErrPiModelNotFound,
	"invalid_model":      ErrPiModelNotFound,
	"protocol":           ErrProviderProtocol,
	"malformed":          ErrProviderProtocol,
}

// messageKeywords is the fallback used when Kind is empty or unrecognized:
// a best-effort substring scan over Message, checked in this order.
var messageKeywords = []struct {
	code     *Error
	keywords []string
}{
	{ErrPiProviderAuth, []string{"unauthorized", "invalid api key", "invalid_api_key", "authentication failed", "forbidden", "401", "403"}},
	{ErrPiProviderQuota, []string{"quota", "rate limit", "rate_limit", "insufficient credits", "429", "too many requests"}},
	{ErrPiModelNotFound, []string{"model not found", "unknown model", "unsupported model", "no such model"}},
	{ErrProviderProtocol, []string{"malformed", "unexpected response", "protocol error", "invalid json", "parse error"}},
}

// TranslateProviderError maps one provider-layer failure Pi reported over
// RPC to a stable Brunel error code. The provider's human-readable message
// is preserved for display, but first passed through redact.Secrets: a
// provider is free to echo an Authorization header or API key back in its
// error text, and spec.md §9 forbids a public error from carrying an API
// key, Authorization header, or unmasked known secret. Classification
// (Kind and keyword scan) runs on the masked message; the keywords are
// error categories, not secrets, so masking does not change the outcome.
//
// It never retries and never suppresses the failure - the caller (Issue
// #9's event loop) surfaces the returned error as-is.
func TranslateProviderError(report ProviderErrorReport) error {
	message := redact.Secrets(report.Message)
	if coded, ok := knownKinds[normalizeKind(report.Kind)]; ok {
		return codeError(coded.Code, message, nil)
	}
	lower := strings.ToLower(message)
	for _, entry := range messageKeywords {
		for _, keyword := range entry.keywords {
			if strings.Contains(lower, keyword) {
				return codeError(entry.code.Code, message, nil)
			}
		}
	}
	return codeError(ErrPiProviderError.Code, message, nil)
}

func normalizeKind(kind string) string {
	return strings.ToLower(strings.TrimSpace(kind))
}
