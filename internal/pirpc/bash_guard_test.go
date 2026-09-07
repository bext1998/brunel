package pirpc

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestSourceNeverSendsBashRPCCommand implements TC-PIRPC-001, the minimum
// test protecting INV-9 (`[FROZEN]`, spec.md §10): internal/pirpc must
// never send a `{"type":"bash"}` RPC command to Pi, since that host-level
// side channel would bypass every Go-side safety and workspace guarantee
// (spec.md §6.1 ADR-002 note, Issue #24 Gate 2). This package does not
// build or send RPC commands yet - that is Issue #9's job - but the guard
// is added now, with the package, so every future change to it (in #9 and
// beyond) is checked automatically; this test's own source is excluded so
// documenting the forbidden literal here does not trip it.
func TestSourceNeverSendsBashRPCCommand(t *testing.T) {
	const forbidden = `"type":"bash"`
	// Also reject the same literal with any amount of whitespace around
	// the colon, since real code is unlikely to hand-format compact JSON.
	const forbiddenSpaced = `"type" : "bash"`

	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("cannot list package directory: %v", err)
	}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") {
			continue
		}
		if entry.Name() == "bash_guard_test.go" {
			continue
		}
		data, err := os.ReadFile(filepath.Join(".", entry.Name()))
		if err != nil {
			t.Fatalf("cannot read %s: %v", entry.Name(), err)
		}
		compact := strings.Join(strings.Fields(string(data)), " ")
		if strings.Contains(string(data), forbidden) || strings.Contains(string(data), forbiddenSpaced) || strings.Contains(compact, `"type": "bash"`) || strings.Contains(compact, `"type":"bash"`) {
			t.Fatalf("%s contains the forbidden bash RPC command literal (INV-9)", entry.Name())
		}
	}
}
