package pirpc

import "strings"

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
	"protocol":           ErrPiProtocol,
	"malformed":          ErrPiProtocol,
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
	{ErrPiProtocol, []string{"malformed", "unexpected response", "protocol error", "invalid json", "parse error"}},
}

// TranslateProviderError maps one provider-layer failure Pi reported over
// RPC to a stable Brunel error code, preserving the original message for
// display. It never retries and never suppresses the failure - the caller
// (Issue #9's event loop) surfaces the returned error as-is.
func TranslateProviderError(report ProviderErrorReport) error {
	if coded, ok := knownKinds[normalizeKind(report.Kind)]; ok {
		return codeError(coded.Code, report.Message, nil)
	}
	lower := strings.ToLower(report.Message)
	for _, entry := range messageKeywords {
		for _, keyword := range entry.keywords {
			if strings.Contains(lower, keyword) {
				return codeError(entry.code.Code, report.Message, nil)
			}
		}
	}
	return codeError(ErrPiProviderError.Code, report.Message, nil)
}

func normalizeKind(kind string) string {
	return strings.ToLower(strings.TrimSpace(kind))
}
