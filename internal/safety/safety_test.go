package safety

import (
	"context"
	"errors"
	"testing"
)

// fakeApprover records every Confirm call so tests can assert exactly when
// (and how many times) Gate is allowed to reach it - only Gate.Decide may
// ever call Approver.Confirm (INV-4).
type fakeApprover struct {
	approve bool
	err     error
	calls   []ApprovalPrompt
}

func (f *fakeApprover) Confirm(_ context.Context, prompt ApprovalPrompt) (bool, error) {
	f.calls = append(f.calls, prompt)
	return f.approve, f.err
}

func TestDecideReadToolsAreAlwaysAutoWithZeroApproverCalls(t *testing.T) {
	approver := &fakeApprover{approve: true}
	gate := NewGate(ModeWorkspace, approver, `C:\ws`)

	for _, tool := range []string{"list_files", "search_text", "read_file", "workspace_diff"} {
		decision, err := gate.Decide(context.Background(), ToolCall{Tool: tool})
		if err != nil {
			t.Fatalf("Decide(%s) error = %v", tool, err)
		}
		if decision.Risk != RiskAuto {
			t.Fatalf("Decide(%s).Risk = %v, want RiskAuto", tool, decision.Risk)
		}
	}
	if len(approver.calls) != 0 {
		t.Fatalf("Approver.Confirm called %d times for AUTO tools, want 0 (AC-9)", len(approver.calls))
	}
}

func TestDecidePreciseWriteToolsAreAutoWithZeroApproverCalls(t *testing.T) {
	approver := &fakeApprover{approve: true}
	gate := NewGate(ModeWorkspace, approver, `C:\ws`)

	for _, tool := range []string{"apply_patch", "create_file", "write_file"} {
		decision, err := gate.Decide(context.Background(), ToolCall{Tool: tool})
		if err != nil {
			t.Fatalf("Decide(%s) error = %v", tool, err)
		}
		if decision.Risk != RiskAuto {
			t.Fatalf("Decide(%s).Risk = %v, want RiskAuto", tool, decision.Risk)
		}
	}
	if len(approver.calls) != 0 {
		t.Fatalf("Approver.Confirm called %d times for precise-write tools, want 0", len(approver.calls))
	}
}

func TestDecideRunPowershellAutoForOrdinaryCommand(t *testing.T) {
	approver := &fakeApprover{approve: true}
	gate := NewGate(ModeWorkspace, approver, `C:\ws`)

	decision, err := gate.Decide(context.Background(), ToolCall{Tool: "run_powershell", Command: "go test ./..."})
	if err != nil {
		t.Fatalf("Decide() error = %v", err)
	}
	if decision.Risk != RiskAuto {
		t.Fatalf("Risk = %v, want RiskAuto", decision.Risk)
	}
	if len(approver.calls) != 0 {
		t.Fatalf("Approver.Confirm called for an ordinary command, want 0 calls")
	}
}

func TestDecideRunPowershellConfirmsRepresentativeCommandsAndApproves(t *testing.T) {
	// One representative command per spec.md §6.2 category.
	commands := []string{
		`Remove-Item -Recurse -Force .\build`,
		`git commit -m "wip"`,
		`git push origin main`,
		`git reset --hard HEAD~1`,
		`npm install left-pad`,
		`Invoke-WebRequest -Uri https://example.com`,
		`curl https://example.com`,
		`Start-Process notepad.exe`,
		`Get-Content C:\Windows\System32\drivers\etc\hosts`,
	}
	for _, cmd := range commands {
		approver := &fakeApprover{approve: true}
		gate := NewGate(ModeWorkspace, approver, `C:\ws`)

		decision, err := gate.Decide(context.Background(), ToolCall{Tool: "run_powershell", Command: cmd})
		if err != nil {
			t.Fatalf("Decide(%q) error = %v", cmd, err)
		}
		if decision.Risk != RiskConfirm {
			t.Fatalf("Decide(%q).Risk = %v, want RiskConfirm", cmd, decision.Risk)
		}
		if len(approver.calls) != 1 {
			t.Fatalf("Decide(%q): Approver.Confirm called %d times, want exactly 1", cmd, len(approver.calls))
		}
		if approver.calls[0].Command != cmd {
			t.Fatalf("prompt.Command = %q, want %q", approver.calls[0].Command, cmd)
		}
		if approver.calls[0].Reason == "" {
			t.Fatalf("Decide(%q): prompt.Reason is empty, want a classification reason", cmd)
		}
	}
}

func TestDecideRunPowershellDeniedLeavesCallerUnableToProceed(t *testing.T) {
	approver := &fakeApprover{approve: false}
	gate := NewGate(ModeWorkspace, approver, `C:\ws`)

	_, err := gate.Decide(context.Background(), ToolCall{Tool: "run_powershell", Command: "git push origin main"})
	if ErrorCode(err) != ErrApprovalDenied.Code {
		t.Fatalf("ErrorCode() = %q, want %q (err=%v)", ErrorCode(err), ErrApprovalDenied.Code, err)
	}
	if len(approver.calls) != 1 {
		t.Fatalf("Approver.Confirm called %d times, want exactly 1", len(approver.calls))
	}
}

func TestDecideRunPowershellApproverErrorIsTreatedAsDenied(t *testing.T) {
	approver := &fakeApprover{err: errors.New("tty closed")}
	gate := NewGate(ModeWorkspace, approver, `C:\ws`)

	_, err := gate.Decide(context.Background(), ToolCall{Tool: "run_powershell", Command: "git push origin main"})
	if ErrorCode(err) != ErrApprovalDenied.Code {
		t.Fatalf("ErrorCode() = %q, want %q (err=%v)", ErrorCode(err), ErrApprovalDenied.Code, err)
	}
}

func TestDecideNoTTYFailsClosedWithoutBlocking(t *testing.T) {
	gate := NewGate(ModeWorkspace, nil, `C:\ws`)

	_, err := gate.Decide(context.Background(), ToolCall{Tool: "run_powershell", Command: "git push origin main"})
	if !errors.Is(err, ErrApprovalRequiredNoTTY) {
		t.Fatalf("err = %v, want ErrApprovalRequiredNoTTY", err)
	}
}

func TestDecideReadonlyModeRejectsMutatingToolsWithoutApproverCall(t *testing.T) {
	approver := &fakeApprover{approve: true}
	gate := NewGate(ModeReadonly, approver, `C:\ws`)

	for _, call := range []ToolCall{
		{Tool: "apply_patch"},
		{Tool: "create_file"},
		{Tool: "write_file"},
		{Tool: "run_powershell", Command: "go test ./..."}, // would be AUTO outside readonly mode
	} {
		_, err := gate.Decide(context.Background(), call)
		if ErrorCode(err) != ErrReadonlyMode.Code {
			t.Fatalf("Decide(%s) ErrorCode() = %q, want %q (err=%v)", call.Tool, ErrorCode(err), ErrReadonlyMode.Code, err)
		}
	}
	if len(approver.calls) != 0 {
		t.Fatalf("Approver.Confirm called %d times in readonly mode, want 0 (readonly is a precondition, not a risk decision)", len(approver.calls))
	}
}

func TestDecideReadonlyModeStillAllowsReadTools(t *testing.T) {
	gate := NewGate(ModeReadonly, nil, `C:\ws`)
	for _, tool := range []string{"list_files", "search_text", "read_file", "workspace_diff"} {
		if _, err := gate.Decide(context.Background(), ToolCall{Tool: tool}); err != nil {
			t.Fatalf("Decide(%s) in readonly mode error = %v, want nil", tool, err)
		}
	}
}

func TestDecideRejectsUnknownTool(t *testing.T) {
	gate := NewGate(ModeWorkspace, nil, `C:\ws`)
	_, err := gate.Decide(context.Background(), ToolCall{Tool: "delete_everything"})
	if ErrorCode(err) != ErrInvalidArgument.Code {
		t.Fatalf("ErrorCode() = %q, want %q (err=%v)", ErrorCode(err), ErrInvalidArgument.Code, err)
	}
}
