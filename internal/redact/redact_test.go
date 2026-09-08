package redact

import (
	"strings"
	"testing"
)

func TestSecretsMasksKnownCredentialForms(t *testing.T) {
	cases := []struct {
		name  string
		in    string
		leak  string // substring that must not survive
		keeps string // non-secret substring that must survive
	}{
		{
			name:  "authorization bearer header",
			in:    "request failed: Authorization: Bearer sk-or-v1-abcdef0123456789 rejected",
			leak:  "abcdef0123456789",
			keeps: "request failed",
		},
		{
			name:  "json api_key field",
			in:    `provider said {"api_key":"sk-live-9f8e7d6c5b4a3210","model":"x"}`,
			leak:  "9f8e7d6c5b4a3210",
			keeps: `"model":"x"`,
		},
		{
			name:  "env style assignment",
			in:    "OPENAI_API_KEY=sk-proj-TOPSECRETVALUE123 was set",
			leak:  "TOPSECRETVALUE123",
			keeps: "OPENAI_API_KEY=",
		},
		{
			name:  "bare api key token",
			in:    "unexpected key sk-or-v1-deadbeefdeadbeef in payload",
			leak:  "deadbeefdeadbeef",
			keeps: "unexpected key",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := Secrets(tc.in)
			if strings.Contains(got, tc.leak) {
				t.Fatalf("Secrets(%q) = %q, still contains secret %q", tc.in, got, tc.leak)
			}
			if !strings.Contains(got, tc.keeps) {
				t.Fatalf("Secrets(%q) = %q, dropped non-secret %q", tc.in, got, tc.keeps)
			}
			if !strings.Contains(got, "[REDACTED]") {
				t.Fatalf("Secrets(%q) = %q, expected a [REDACTED] marker", tc.in, got)
			}
		})
	}
}

func TestSecretsLeavesPlainTextUnchanged(t *testing.T) {
	in := "model server returned HTTP 503 after 2 attempts"
	if got := Secrets(in); got != in {
		t.Fatalf("Secrets(%q) = %q, want unchanged", in, got)
	}
}

func TestSecretsRedactsKnownValuesTheHeuristicsMiss(t *testing.T) {
	cases := []struct {
		name string
		in   string
		key  string
	}{
		{"google AIza key", "auth failed for AIzaSyD-ExampleKey-000111222333444 today", "AIzaSyD-ExampleKey-000111222333444"},
		{"groq gsk_ key", "invalid token gsk_ExampleExampleExample000111 supplied", "gsk_ExampleExampleExample000111"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := Secrets(tc.in); !strings.Contains(got, tc.key) {
				t.Fatalf("Secrets(..) = %q, heuristics unexpectedly caught %q; fold it into the regex set and drop this case", got, tc.key)
			}
			got := Secrets(tc.in, tc.key)
			if strings.Contains(got, tc.key) {
				t.Fatalf("Secrets(.., known) = %q, known value %q not redacted", got, tc.key)
			}
			if !strings.Contains(got, "[REDACTED]") {
				t.Fatalf("Secrets(.., known) = %q, expected a [REDACTED] marker", got)
			}
		})
	}
}

func TestSecretsIgnoresEmptyOrShortKnownValues(t *testing.T) {
	in := "provider returned an opaque failure"
	// "" and short values must not turn into a [REDACTED] carpet.
	if got := Secrets(in, "", "abc", "  "); got != in {
		t.Fatalf("Secrets(%q, empty/short...) = %q, want unchanged", in, got)
	}
}
