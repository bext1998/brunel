package pirpc

import "testing"

func TestTranslateProviderErrorByKind(t *testing.T) {
	cases := map[string]string{
		"auth":            ErrPiProviderAuth.Code,
		"authentication":  ErrPiProviderAuth.Code,
		"unauthorized":    ErrPiProviderAuth.Code,
		"quota":           ErrPiProviderQuota.Code,
		"rate_limit":      ErrPiProviderQuota.Code,
		"model_not_found": ErrPiModelNotFound.Code,
		"invalid_model":   ErrPiModelNotFound.Code,
		"protocol":        ErrPiProtocol.Code,
		"malformed":       ErrPiProtocol.Code,
	}
	for kind, want := range cases {
		err := TranslateProviderError(ProviderErrorReport{Kind: kind, Message: "some message"})
		if ErrorCode(err) != want {
			t.Errorf("TranslateProviderError(Kind=%q) = %q, want %q", kind, ErrorCode(err), want)
		}
	}
}

func TestTranslateProviderErrorFallsBackToMessageKeywords(t *testing.T) {
	cases := map[string]string{
		"401 Unauthorized: invalid api key":        ErrPiProviderAuth.Code,
		"rate limit exceeded, try again later":     ErrPiProviderQuota.Code,
		"model not found: gpt-9000":                ErrPiModelNotFound.Code,
		"protocol error: unexpected response type": ErrPiProtocol.Code,
	}
	for message, want := range cases {
		err := TranslateProviderError(ProviderErrorReport{Message: message})
		if ErrorCode(err) != want {
			t.Errorf("TranslateProviderError(Message=%q) = %q, want %q", message, ErrorCode(err), want)
		}
	}
}

func TestTranslateProviderErrorDefaultsToGenericProviderError(t *testing.T) {
	err := TranslateProviderError(ProviderErrorReport{Message: "the model server had a hiccup"})
	if ErrorCode(err) != ErrPiProviderError.Code {
		t.Fatalf("ErrorCode() = %q, want %q", ErrorCode(err), ErrPiProviderError.Code)
	}
}

func TestTranslateProviderErrorPreservesMessage(t *testing.T) {
	err := TranslateProviderError(ProviderErrorReport{Kind: "quota", Message: "insufficient credits on account acct_123"})
	if err.Error() != ErrPiProviderQuota.Code+": insufficient credits on account acct_123" {
		t.Fatalf("err.Error() = %q, want message preserved", err.Error())
	}
}

func TestTranslateProviderErrorNeverRetriesJustClassifies(t *testing.T) {
	// TranslateProviderError has no retry/backoff of its own: calling it
	// repeatedly with the same report always yields the same classification.
	report := ProviderErrorReport{Kind: "quota", Message: "rate limited"}
	first := TranslateProviderError(report)
	second := TranslateProviderError(report)
	if ErrorCode(first) != ErrorCode(second) {
		t.Fatalf("classification is not stable across calls: %q vs %q", ErrorCode(first), ErrorCode(second))
	}
}
