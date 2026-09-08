package pirpc

import (
	"reflect"
	"testing"
)

func TestBuildArgsMatchesFrozenCommandLineWithProvider(t *testing.T) {
	args, err := BuildArgs(LaunchOptions{Provider: "openrouter", Model: "anthropic/claude-sonnet-4"})
	if err != nil {
		t.Fatalf("BuildArgs() error = %v", err)
	}
	want := []string{
		"--mode", "rpc",
		"--no-builtin-tools",
		"--no-extensions",
		"-e", "taylor-tools.ts",
		"--no-session",
		"--provider", "openrouter",
		"--model", "anthropic/claude-sonnet-4",
	}
	if !reflect.DeepEqual(args, want) {
		t.Fatalf("BuildArgs() = %v, want %v", args, want)
	}
}

func TestBuildArgsOmitsProviderFlagWhenEmpty(t *testing.T) {
	args, err := BuildArgs(LaunchOptions{Model: "openrouter/anthropic/claude-sonnet-4"})
	if err != nil {
		t.Fatalf("BuildArgs() error = %v", err)
	}
	for _, a := range args {
		if a == "--provider" {
			t.Fatalf("BuildArgs() included --provider when Provider was empty: %v", args)
		}
	}
	if args[len(args)-2] != "--model" || args[len(args)-1] != "openrouter/anthropic/claude-sonnet-4" {
		t.Fatalf("BuildArgs() did not end with --model <m>: %v", args)
	}
}

func TestBuildArgsUsesCustomExtensionPath(t *testing.T) {
	args, err := BuildArgs(LaunchOptions{Model: "m", ExtensionPath: `C:\brunel\taylor-tools.ts`})
	if err != nil {
		t.Fatalf("BuildArgs() error = %v", err)
	}
	found := false
	for i, a := range args {
		if a == "-e" && i+1 < len(args) && args[i+1] == `C:\brunel\taylor-tools.ts` {
			found = true
		}
	}
	if !found {
		t.Fatalf("BuildArgs() did not use custom extension path: %v", args)
	}
}

func TestEffectiveProvider(t *testing.T) {
	cases := []struct {
		name string
		opts LaunchOptions
		want string
	}{
		{"explicit provider wins over model prefix", LaunchOptions{Provider: "openrouter", Model: "anthropic/claude-sonnet-4"}, "openrouter"},
		{"model prefix used when provider empty", LaunchOptions{Model: "openrouter/anthropic/claude-sonnet-4"}, "openrouter"},
		{"single-segment provider prefix", LaunchOptions{Model: "openai/gpt-4o"}, "openai"},
		{"no provider and no prefix", LaunchOptions{Model: "gpt-4o"}, ""},
		{"whitespace is trimmed", LaunchOptions{Model: "  gemini/pro  "}, "gemini"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.opts.EffectiveProvider(); got != tc.want {
				t.Fatalf("EffectiveProvider() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestBuildArgsRejectsEmptyModel(t *testing.T) {
	if _, err := BuildArgs(LaunchOptions{Provider: "openrouter"}); ErrorCode(err) != ErrInvalidArgument.Code {
		t.Fatalf("ErrorCode() = %q, want %q (err=%v)", ErrorCode(err), ErrInvalidArgument.Code, err)
	}
	if _, err := BuildArgs(LaunchOptions{Model: "   "}); ErrorCode(err) != ErrInvalidArgument.Code {
		t.Fatalf("ErrorCode() = %q, want %q for whitespace-only model", ErrorCode(err), ErrInvalidArgument.Code)
	}
}
