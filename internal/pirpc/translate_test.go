package pirpc

import (
	"strings"
	"testing"
)

func TestTranslateProviderErrorByKind(t *testing.T) {
	cases := map[string]string{
		"auth":            ErrPiProviderAuth.Code,
		"authentication":  ErrPiProviderAuth.Code,
		"unauthorized":    ErrPiProviderAuth.Code,
		"quota":           ErrPiProviderQuota.Code,
		"rate_limit":      ErrPiProviderQuota.Code,
		"model_not_found": ErrPiModelNotFound.Code,
		"invalid_model":   ErrPiModelNotFound.Code,
		"protocol":        ErrProviderProtocol.Code,
		"malformed":       ErrProviderProtocol.Code,
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
		"protocol error: unexpected response type": ErrProviderProtocol.Code,
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

func TestTranslateProviderErrorUsesSpecMandatedProtocolCode(t *testing.T) {
	// spec.md §11 EC-11 fixes the public code for a malformed/unexpected
	// provider protocol response as E_PROVIDER_PROTOCOL (not Pi-prefixed).
	err := TranslateProviderError(ProviderErrorReport{Kind: "protocol", Message: "bad frame"})
	if ErrorCode(err) != "E_PROVIDER_PROTOCOL" {
		t.Fatalf("ErrorCode() = %q, want E_PROVIDER_PROTOCOL (spec.md EC-11)", ErrorCode(err))
	}
}

func TestTranslateProviderErrorMasksCredentialsInMessage(t *testing.T) {
	// A provider is free to echo the offending request back in its error
	// text. spec.md §9: a public error must not carry an API key,
	// Authorization header, or unmasked known secret.
	cases := []struct {
		name string
		in   string
		leak string
	}{
		{
			name: "authorization bearer header",
			in:   "upstream 401: Authorization: Bearer sk-or-v1-0123456789abcdef rejected",
			leak: "sk-or-v1-0123456789abcdef",
		},
		{
			name: "echoed api key assignment",
			in:   "provider rejected api key OPENROUTER_API_KEY=sk-or-v1-SECRETSECRETSECRET: unauthorized",
			leak: "sk-or-v1-SECRETSECRETSECRET",
		},
		{
			name: "json api_key field",
			in:   `provider error {"api_key":"sk-live-abcd1234efgh5678","code":"invalid_api_key"}`,
			leak: "sk-live-abcd1234efgh5678",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := TranslateProviderError(ProviderErrorReport{Message: tc.in})
			if got := err.Error(); strings.Contains(got, tc.leak) {
				t.Fatalf("translated error %q still contains secret %q", got, tc.leak)
			}
			// The auth classification must still work off the masked text.
			if ErrorCode(err) != ErrPiProviderAuth.Code {
				t.Fatalf("ErrorCode() = %q, want %q", ErrorCode(err), ErrPiProviderAuth.Code)
			}
		})
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
