package pirpc

import (
	"strings"
	"testing"
)

func TestCredentialEnvVarKnownProviders(t *testing.T) {
	cases := map[string]string{
		"openrouter": "OPENROUTER_API_KEY",
		"OpenRouter": "OPENROUTER_API_KEY",
		"anthropic":  "ANTHROPIC_API_KEY",
		"openai":     "OPENAI_API_KEY",
		"gemini":     "GEMINI_API_KEY",
		"google":     "GEMINI_API_KEY",
	}
	for provider, want := range cases {
		got, ok := CredentialEnvVar(provider)
		if !ok || got != want {
			t.Errorf("CredentialEnvVar(%q) = (%q, %v), want (%q, true)", provider, got, ok, want)
		}
	}
}

func TestCredentialEnvVarUnknownProvider(t *testing.T) {
	if _, ok := CredentialEnvVar("some-future-provider"); ok {
		t.Error("CredentialEnvVar() unexpectedly recognized an unknown provider")
	}
}

func TestInjectCredentialsSetsNewVariable(t *testing.T) {
	base := []string{"PATH=C:\\Windows", "TEMP=C:\\Temp"}
	out, err := InjectCredentials(base, "openrouter", OpenRouterCredential("sk-test-key"))
	if err != nil {
		t.Fatalf("InjectCredentials() error = %v", err)
	}

	if len(out) != 3 {
		t.Fatalf("len(out) = %d, want 3", len(out))
	}
	if !contains(out, "OPENROUTER_API_KEY=sk-test-key") {
		t.Fatalf("out = %v, missing injected credential", out)
	}
	// base must not be mutated.
	if len(base) != 2 {
		t.Fatalf("base was mutated: %v", base)
	}
}

func TestInjectCredentialsReplacesExistingVariable(t *testing.T) {
	base := []string{"OPENROUTER_API_KEY=old-leaked-value", "PATH=C:\\Windows"}
	out, err := InjectCredentials(base, "openrouter", OpenRouterCredential("sk-fresh-key"))
	if err != nil {
		t.Fatalf("InjectCredentials() error = %v", err)
	}

	if len(out) != 2 {
		t.Fatalf("len(out) = %d, want 2 (replace, not append)", len(out))
	}
	if !contains(out, "OPENROUTER_API_KEY=sk-fresh-key") {
		t.Fatalf("out = %v, want replaced credential", out)
	}
	if contains(out, "OPENROUTER_API_KEY=old-leaked-value") {
		t.Fatalf("out = %v, old credential value leaked through", out)
	}
}

func TestInjectCredentialsRejectsProviderMismatch(t *testing.T) {
	// The defect this guards: a caller passing the OpenRouter key that
	// config resolves today under provider="openai" would otherwise put an
	// OpenRouter key into OPENAI_API_KEY.
	base := []string{"PATH=C:\\Windows"}
	out, err := InjectCredentials(base, "openai", OpenRouterCredential("sk-or-v1-secretvalue"))
	if ErrorCode(err) != ErrCredentialProviderMismatch.Code {
		t.Fatalf("ErrorCode(err) = %q, want %q", ErrorCode(err), ErrCredentialProviderMismatch.Code)
	}
	if len(out) != 1 || out[0] != "PATH=C:\\Windows" {
		t.Fatalf("out = %v, want environment unchanged on mismatch", out)
	}
	for _, e := range out {
		if strings.Contains(e, "sk-or-v1-secretvalue") {
			t.Fatalf("out = %v, mismatched key leaked into environment", out)
		}
	}
}

func TestInjectCredentialsAcceptsProviderAlias(t *testing.T) {
	// "google" and "gemini" resolve to the same Pi variable, so a key
	// resolved for one may be injected when launching the other.
	base := []string{"PATH=C:\\Windows"}
	out, err := InjectCredentials(base, "gemini", Credential{Provider: "google", APIKey: "sk-gemini-key"})
	if err != nil {
		t.Fatalf("InjectCredentials() error = %v", err)
	}
	if !contains(out, "GEMINI_API_KEY=sk-gemini-key") {
		t.Fatalf("out = %v, want alias-matched credential injected", out)
	}
}

func TestInjectCredentialsLeavesEnvironmentUnchangedForUnknownProvider(t *testing.T) {
	base := []string{"PATH=C:\\Windows"}
	out, err := InjectCredentials(base, "some-future-provider", Credential{Provider: "some-future-provider", APIKey: "sk-test-key"})
	if err != nil {
		t.Fatalf("InjectCredentials() error = %v", err)
	}
	if len(out) != 1 || out[0] != "PATH=C:\\Windows" {
		t.Fatalf("out = %v, want environment unchanged for an unrecognized provider", out)
	}
}

func TestInjectCredentialsLeavesEnvironmentUnchangedForEmptyKey(t *testing.T) {
	base := []string{"PATH=C:\\Windows"}
	out, err := InjectCredentials(base, "openrouter", OpenRouterCredential(""))
	if err != nil {
		t.Fatalf("InjectCredentials() error = %v", err)
	}
	if len(out) != 1 || out[0] != "PATH=C:\\Windows" {
		t.Fatalf("out = %v, want environment unchanged for an empty key", out)
	}
}

func contains(entries []string, want string) bool {
	for _, e := range entries {
		if e == want {
			return true
		}
	}
	return false
}
