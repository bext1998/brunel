package pirpc

import "strings"

// defaultExtensionPath matches the literal example in spec.md §5.1: the
// taylor-tools.ts extension is expected next to the working directory
// unless the caller (Issue #9, which actually starts the subprocess)
// resolves and overrides it with an absolute path.
const defaultExtensionPath = "taylor-tools.ts"

// LaunchOptions is what #8 must pass through to the Pi RPC subprocess
// (spec.md §5.3): the user-selected model, in whatever provider-prefixed
// syntax Pi's own --model flag expects, and an optional explicit provider.
// Neither field's syntax is validated beyond "non-empty" - Brunel does not
// claim to know or enforce Pi's own provider/model naming conventions
// (spec.md §5.3: "不得由 Brunel 片面承諾涵蓋範圍").
type LaunchOptions struct {
	// Model is required and passed verbatim as pi's --model argument.
	Model string
	// Provider is optional. When set, it is passed as pi's --provider
	// argument ahead of --model, matching the literal launch command in
	// spec.md §5.1. When empty, no --provider flag is added at all -
	// Pi's own --model syntax may already embed a provider prefix.
	Provider string
	// ExtensionPath overrides the taylor-tools.ts path passed to `-e`.
	// Empty uses defaultExtensionPath.
	ExtensionPath string
}

// EffectiveProvider is the single place that answers "which provider do
// these options actually target": the explicit Provider when set, otherwise
// the prefix of Model up to the first "/" (Pi's --model convention is
// "<provider>/<model...>" when no --provider flag is given, per the launch
// command in spec.md §5.1). It returns "" only when neither is available.
//
// Credential injection must use this, not opts.Provider directly:
// BuildArgs already handles a provider-prefixed model by passing it through
// verbatim, so a caller is allowed to leave Provider empty, and reading
// opts.Provider alone would then miss the Credential Manager key entirely.
func (opts LaunchOptions) EffectiveProvider() string {
	if p := strings.TrimSpace(opts.Provider); p != "" {
		return p
	}
	if prefix, _, ok := strings.Cut(strings.TrimSpace(opts.Model), "/"); ok {
		return strings.TrimSpace(prefix)
	}
	return ""
}

// BuildArgs returns the full argument list for launching `pi` as an RPC
// subprocess, matching the literal frozen command in spec.md §5.1:
//
//	pi --mode rpc --no-builtin-tools --no-extensions -e <ext> --no-session [--provider <p>] --model <m>
//
// It only builds the argument list - it does not start the process. That,
// and the actual RPC event loop, is Issue #9's responsibility; #8 is
// limited to passthrough and translation (spec.md's own scope for
// TASK-A1-F07).
func BuildArgs(opts LaunchOptions) ([]string, error) {
	model := strings.TrimSpace(opts.Model)
	if model == "" {
		return nil, codeError(ErrInvalidArgument.Code, "model is required", nil)
	}
	extension := strings.TrimSpace(opts.ExtensionPath)
	if extension == "" {
		extension = defaultExtensionPath
	}

	args := []string{
		"--mode", "rpc",
		"--no-builtin-tools",
		"--no-extensions",
		"-e", extension,
		"--no-session",
	}
	if provider := strings.TrimSpace(opts.Provider); provider != "" {
		args = append(args, "--provider", provider)
	}
	args = append(args, "--model", model)
	return args, nil
}
