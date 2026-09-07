package pirpc

import "testing"

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
	out := InjectCredentials(base, "openrouter", "sk-test-key")

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
	out := InjectCredentials(base, "openrouter", "sk-fresh-key")

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

func TestInjectCredentialsLeavesEnvironmentUnchangedForUnknownProvider(t *testing.T) {
	base := []string{"PATH=C:\\Windows"}
	out := InjectCredentials(base, "some-future-provider", "sk-test-key")
	if len(out) != 1 || out[0] != "PATH=C:\\Windows" {
		t.Fatalf("out = %v, want environment unchanged for an unrecognized provider", out)
	}
}

func TestInjectCredentialsLeavesEnvironmentUnchangedForEmptyKey(t *testing.T) {
	base := []string{"PATH=C:\\Windows"}
	out := InjectCredentials(base, "openrouter", "")
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
