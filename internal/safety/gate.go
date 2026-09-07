package safety

import (
	"context"
	"fmt"
)

// Risk is the outcome of classifying one tool call. [FROZEN] (spec.md §6.2).
type Risk int

const (
	RiskAuto Risk = iota
	RiskConfirm
)

func (r Risk) String() string {
	switch r {
	case RiskAuto:
		return "auto"
	case RiskConfirm:
		return "confirm"
	default:
		return "unknown"
	}
}

// ApprovalPrompt is the UI-neutral content shown for a CONFIRM decision.
// [FROZEN] (spec.md §5.2).
type ApprovalPrompt struct {
	Command string
	Reason  string
}

// Approver asks a human to approve or reject one CONFIRM decision.
// [FROZEN] (spec.md §5.2). The TUI and any TTY-backed plain text mode each
// implement this; without a TTY no Approver is provided at all.
type Approver interface {
	Confirm(ctx context.Context, prompt ApprovalPrompt) (bool, error)
}

// Mode selects whether mutating tools are even considered.
type Mode int

const (
	// ModeWorkspace is the default: every one of the 8 tools may run,
	// subject to the usual AUTO/CONFIRM classification.
	ModeWorkspace Mode = iota
	// ModeReadonly deterministically rejects every mutating tool call
	// before classification, without ever invoking Approver (spec.md §6.3).
	ModeReadonly
)

// mutatingTools is the subset of the 8 frozen tools (spec.md §5.4) that
// perform writes or run arbitrary commands, and are therefore rejected
// outright in readonly mode.
var mutatingTools = map[string]bool{
	"apply_patch":    true,
	"create_file":    true,
	"write_file":     true,
	"run_powershell": true,
}

// readonlyTools is the subset allowed in readonly mode (spec.md §4.2/CT-1).
var readonlyTools = map[string]bool{
	"list_files":     true,
	"search_text":    true,
	"read_file":      true,
	"workspace_diff": true,
}

// ToolCall is the minimal information Decide needs to classify one call to
// one of the 8 frozen tools. Command is only meaningful for
// run_powershell; every other tool is always AUTO (spec.md §6.2: "精確寫入"
// - exact, hash-guarded writes - are themselves already AUTO).
type ToolCall struct {
	Tool    string
	Command string
}

// Decision is the result of a successful Decide call: the call is allowed
// to proceed, and Risk/Reason are recorded for the caller's own event log
// or transcript.
type Decision struct {
	Risk   Risk
	Reason string
}

// Gate is Brunel's single safety decision entry point (INV-1, INV-4): every
// one of the 8 tools must call Decide before any I/O, and only Gate may
// call Approver.Confirm.
type Gate struct {
	Mode          Mode
	Approver      Approver // nil when there is no TTY
	WorkspaceRoot string   // absolute, resolved workspace root; used to flag out-of-workspace absolute paths in command text
}

// NewGate constructs a Gate. approver may be nil (no TTY): Decide then
// fails any CONFIRM decision with ErrApprovalRequiredNoTTY instead of
// blocking.
func NewGate(mode Mode, approver Approver, workspaceRoot string) *Gate {
	return &Gate{Mode: mode, Approver: approver, WorkspaceRoot: workspaceRoot}
}

// Decide is the single safety decision entry point. It never has a side
// effect of its own: it only decides whether the caller may proceed to
// perform the tool's actual I/O.
//
//   - An unknown tool name is rejected outright (defense in depth: only the
//     8 frozen tools may ever reach a decision).
//   - In ModeReadonly, any mutating tool is rejected before classification
//     (a tool precondition, not a risk decision - spec.md §6.3) - Approver
//     is never invoked and there is no confirmation prompt.
//   - Otherwise the call is classified AUTO or CONFIRM (classify.go).
//     AUTO returns immediately with zero Approver calls (AC-9). CONFIRM
//     calls g.Approver.Confirm once; with no Approver at all (no TTY) it
//     fails closed with ErrApprovalRequiredNoTTY instead of executing
//     (spec.md §6.3, EC-6) - and a decline fails with ErrApprovalDenied.
//
// Every path - rejection, denial, or no-TTY - returns before any tool I/O
// has happened; only a nil error means the caller may proceed.
func (g *Gate) Decide(ctx context.Context, call ToolCall) (Decision, error) {
	if !readonlyTools[call.Tool] && !mutatingTools[call.Tool] {
		return Decision{}, codeError(ErrInvalidArgument.Code, fmt.Sprintf("unknown tool %q", call.Tool), nil)
	}

	if g.Mode == ModeReadonly && mutatingTools[call.Tool] {
		return Decision{}, codeError(ErrReadonlyMode.Code, fmt.Sprintf("tool %q is rejected in readonly mode", call.Tool), nil)
	}

	risk, reason := g.classify(call)
	if risk == RiskAuto {
		return Decision{Risk: RiskAuto}, nil
	}

	if g.Approver == nil {
		return Decision{}, ErrApprovalRequiredNoTTY
	}
	approved, err := g.Approver.Confirm(ctx, ApprovalPrompt{Command: call.Command, Reason: reason})
	if err != nil {
		return Decision{}, codeError(ErrApprovalDenied.Code, "approver failed", err)
	}
	if !approved {
		return Decision{}, codeError(ErrApprovalDenied.Code, "user declined confirmation", nil)
	}
	return Decision{Risk: RiskConfirm, Reason: reason}, nil
}
