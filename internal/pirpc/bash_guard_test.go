package pirpc

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestSourceHasNoLiteralBashRPCCommand is an interim, best-effort smoke
// check for INV-9 (`[FROZEN]`, spec.md §10): internal/pirpc must never send
// a `{"type":"bash"}` RPC command to Pi, since that host-level side channel
// would bypass every Go-side safety and workspace guarantee (spec.md §6.1
// ADR-002 note, Issue #24 Gate 2).
//
// This is NOT the full guard and does not by itself satisfy INV-9. It only
// scans package source for the forbidden JSON literal in a few common
// spellings. It would miss, for example, a command built with
// json.Marshal over a struct whose field is tagged `json:"type"` and set to
// "bash", or one assembled by string concatenation. The complete
// AST/lint check plus a dedicated CI stage - and the formal TC-PIRPC-001
// positive/negative cases - are tracked in Issue #30; until that lands,
// treat this test as a low-cost regression tripwire, not proof.
//
// This package does not build or send RPC commands yet (that is Issue #9's
// job); the tripwire is added now so an accidental literal in later changes
// is caught early. This test's own source is excluded so documenting the
// forbidden literal here does not trip it.
func TestSourceHasNoLiteralBashRPCCommand(t *testing.T) {
	// The literal in the spellings a hand-written command is most likely
	// to use: raw compact JSON, raw JSON with spaces around the colon, and
	// the same inside a double-quoted (escaped) Go string literal.
	forbidden := []string{
		`"type":"bash"`,
		`"type": "bash"`,
		`"type" : "bash"`,
		`\"type\":\"bash\"`,
		`\"type\": \"bash\"`,
	}

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
		raw := string(data)
		compact := strings.Join(strings.Fields(raw), " ")
		for _, needle := range forbidden {
			if strings.Contains(raw, needle) || strings.Contains(compact, needle) {
				t.Fatalf("%s contains a literal bash RPC command spelling %q (INV-9; full guard tracked in Issue #30)", entry.Name(), needle)
			}
		}
	}
}
