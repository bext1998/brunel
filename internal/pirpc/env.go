package pirpc

import "strings"

// providerCredentialEnv maps a provider name to the environment variable
// Pi itself already recognizes for that provider's API key (confirmed
// against @earendil-works/pi-coding-agent by Issue #24's compatibility
// spike, spikes/pi-compatibility/recon-credentials.ps1 on the
// agent/pi-spike-issue-24 branch). Brunel injects into this existing
// mechanism rather than inventing its own (spec.md §5.3: "透過環境變數或 Pi
// 既有的 credential 機制傳入"). Names are matched case-insensitively.
var providerCredentialEnv = map[string]string{
	"openrouter": "OPENROUTER_API_KEY",
	"anthropic":  "ANTHROPIC_API_KEY",
	"openai":     "OPENAI_API_KEY",
	"gemini":     "GEMINI_API_KEY",
	"google":     "GEMINI_API_KEY",
}

// CredentialEnvVar returns the environment variable name Pi recognizes for
// provider's API key, and whether provider is one Brunel knows about.
// Brunel's own credential storage today (Issue #12) only ever resolves an
// OpenRouter key; other providers are listed here so a caller with its own
// key source can still inject correctly, and so this mapping has one place
// to grow as Brunel's config layer gains more providers.
func CredentialEnvVar(provider string) (string, bool) {
	name, ok := providerCredentialEnv[strings.ToLower(strings.TrimSpace(provider))]
	return name, ok
}

// InjectCredentials returns a copy of base (an os.Environ()-shaped slice of
// "KEY=VALUE" strings) with provider's credential environment variable set
// to apiKey, replacing any existing entry for that variable. base is never
// modified in place. An unrecognized provider or an empty apiKey leaves the
// environment unchanged - the subprocess then falls back to whatever
// credential Pi can find on its own (its settings.json, or a variable the
// user already has set), which is an accepted limitation for providers
// Brunel's own config layer does not yet resolve a key for (spec.md §5.3:
// "須在 README／CLI help 明確揭露為「隨 Pi 版本變動」").
func InjectCredentials(base []string, provider, apiKey string) []string {
	name, ok := CredentialEnvVar(provider)
	if !ok || strings.TrimSpace(apiKey) == "" {
		out := make([]string, len(base))
		copy(out, base)
		return out
	}

	out := make([]string, 0, len(base)+1)
	prefix := name + "="
	replaced := false
	for _, entry := range base {
		if strings.HasPrefix(entry, prefix) {
			out = append(out, prefix+apiKey)
			replaced = true
			continue
		}
		out = append(out, entry)
	}
	if !replaced {
		out = append(out, prefix+apiKey)
	}
	return out
}
