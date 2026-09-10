//go:build windows

package tools

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	osexec "os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	brunelexec "github.com/bext1998/brunel/internal/exec"
	"github.com/bext1998/brunel/internal/safety"
	"github.com/bext1998/brunel/internal/workspace"
)

type fakeApprover struct {
	approve bool
	calls   int
}

func (a *fakeApprover) Confirm(context.Context, safety.ApprovalPrompt) (bool, error) {
	a.calls++
	return a.approve, nil
}

func TestRegistryRejectsUnknownAndInvalidParamsWithoutSideEffects(t *testing.T) {
	root := t.TempDir()
	writeTestFile(t, filepath.Join(root, "note.txt"), "original\n")
	registry := newTestRegistry(t, root, safety.ModeWorkspace, nil, nil)

	if _, err := registry.Call(context.Background(), "not_a_tool", json.RawMessage(`{}`)); !errors.Is(err, ErrInvalidArgument) {
		t.Fatalf("unknown tool error = %v, want %v", err, ErrInvalidArgument)
	}
	if _, err := registry.Call(context.Background(), "write_file", json.RawMessage(`{"path":"note.txt","expected_hash":"hash","content":"changed","extra":true}`)); !errors.Is(err, ErrInvalidArgument) {
		t.Fatalf("unknown parameter error = %v, want %v", err, ErrInvalidArgument)
	}
	if _, err := registry.Call(context.Background(), "write_file", json.RawMessage(`{"path":"note.txt","expected_hash":"hash","content":1}`)); !errors.Is(err, ErrInvalidArgument) {
		t.Fatalf("wrong-type parameter error = %v, want %v", err, ErrInvalidArgument)
	}
	if _, err := registry.Call(context.Background(), "write_file", json.RawMessage(`{"path":"note.txt","content":"changed"}`)); !errors.Is(err, ErrInvalidArgument) {
		t.Fatalf("missing expected_hash error = %v, want %v", err, ErrInvalidArgument)
	}
	if _, err := registry.Call(context.Background(), "apply_patch", json.RawMessage(`{"path":"note.txt","hunks":[]}`)); !errors.Is(err, ErrInvalidArgument) {
		t.Fatalf("missing patch expected_hash error = %v, want %v", err, ErrInvalidArgument)
	}
	assertFileContent(t, filepath.Join(root, "note.txt"), "original\n")
}

func TestRegistryDeniedPowerShellHasNoBypass(t *testing.T) {
	root := t.TempDir()
	writeTestFile(t, filepath.Join(root, "x"), "present\n")
	approver := &fakeApprover{approve: false}
	registry := newTestRegistry(t, root, safety.ModeWorkspace, approver, nil)

	_, err := registry.Call(context.Background(), "run_powershell", json.RawMessage(`{"command":"Remove-Item .\\x -Recurse"}`))
	if !errors.Is(err, safety.ErrApprovalDenied) {
		t.Fatalf("denied command error = %v, want %v", err, safety.ErrApprovalDenied)
	}
	if approver.calls != 1 {
		t.Fatalf("approver calls = %d, want 1", approver.calls)
	}
	assertFileContent(t, filepath.Join(root, "x"), "present\n")
}

func TestRegistryReadonlyRejectsMutationsAndAllowsReads(t *testing.T) {
	if _, err := osexec.LookPath("git"); err != nil {
		t.Skip("git is required for workspace_diff: ", err)
	}
	root := t.TempDir()
	writeTestFile(t, filepath.Join(root, "note.txt"), "needle\n")
	gitRun(t, root, "init")
	registry := newTestRegistry(t, root, safety.ModeReadonly, nil, nil)

	mutations := []struct {
		name   string
		params string
	}{
		{"create_file", `{"path":"new.txt","content":"new"}`},
		{"write_file", `{"path":"note.txt","expected_hash":"hash","content":"changed"}`},
		{"apply_patch", `{"path":"note.txt","expected_hash":"hash","hunks":[]}`},
		{"run_powershell", `{"command":"Write-Output blocked"}`},
	}
	for _, call := range mutations {
		if _, err := registry.Call(context.Background(), call.name, json.RawMessage(call.params)); !errors.Is(err, safety.ErrReadonlyMode) {
			t.Fatalf("%s readonly error = %v, want %v", call.name, err, safety.ErrReadonlyMode)
		}
	}
	assertFileContent(t, filepath.Join(root, "note.txt"), "needle\n")
	if _, err := os.Stat(filepath.Join(root, "new.txt")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("readonly create_file created new.txt: %v", err)
	}

	if result, err := registry.Call(context.Background(), "list_files", json.RawMessage(`{"path":"."}`)); err != nil || len(result.List.Entries) == 0 {
		t.Fatalf("readonly list_files result=%#v err=%v", result, err)
	}
	if result, err := registry.Call(context.Background(), "search_text", json.RawMessage(`{"pattern":"needle"}`)); err != nil || len(result.Search.Matches) != 1 {
		t.Fatalf("readonly search_text result=%#v err=%v", result, err)
	}
	if result, err := registry.Call(context.Background(), "read_file", json.RawMessage(`{"path":"note.txt"}`)); err != nil || result.ReadFile.Hash == "" {
		t.Fatalf("readonly read_file result=%#v err=%v", result, err)
	}
	if _, err := registry.Call(context.Background(), "workspace_diff", json.RawMessage(`{}`)); err != nil {
		t.Fatalf("readonly workspace_diff error = %v", err)
	}
}

func TestRegistryRejectsPathEscape(t *testing.T) {
	root := t.TempDir()
	registry := newTestRegistry(t, root, safety.ModeWorkspace, nil, nil)
	for _, path := range []string{`..\\outside`, `C:\\Windows\\x`} {
		params, err := json.Marshal(ReadFileParams{Path: path})
		if err != nil {
			t.Fatalf("marshal params: %v", err)
		}
		if _, err := registry.Call(context.Background(), "read_file", params); !errors.Is(err, workspace.ErrPathEscape) {
			t.Fatalf("path %q error = %v, want %v", path, err, workspace.ErrPathEscape)
		}
	}
}

func TestAC6ToolClosedLoop(t *testing.T) {
	if _, err := osexec.LookPath("git"); err != nil {
		t.Skip("git is required for AC-6: ", err)
	}
	runner, err := brunelexec.NewRunner()
	if err != nil {
		t.Skip("pwsh is required for AC-6: ", err)
	}
	root := t.TempDir()
	writeTestFile(t, filepath.Join(root, "seed.txt"), "needle old\n")
	gitRun(t, root, "init")
	gitRun(t, root, "add", "seed.txt")
	registry := newTestRegistry(t, root, safety.ModeWorkspace, nil, runner)

	search, err := registry.Call(context.Background(), "search_text", json.RawMessage(`{"pattern":"needle"}`))
	if err != nil || len(search.Search.Matches) != 1 || search.Search.Matches[0].Path != "seed.txt" {
		t.Fatalf("search result=%#v err=%v", search, err)
	}
	read, err := registry.Call(context.Background(), "read_file", json.RawMessage(`{"path":"seed.txt"}`))
	if err != nil || read.ReadFile.Hash == "" || len(read.ReadFile.Lines) != 1 || read.ReadFile.Lines[0].Text != "needle old" {
		t.Fatalf("read result=%#v err=%v", read, err)
	}
	patchParams, err := json.Marshal(ApplyPatchParams{
		Path:         "seed.txt",
		ExpectedHash: read.ReadFile.Hash,
		Hunks:        []PatchHunk{{StartLine: 1, EndLine: 1, OldLines: []string{"needle old"}, NewLines: []string{"needle new"}}},
	})
	if err != nil {
		t.Fatalf("marshal patch params: %v", err)
	}
	patch, err := registry.Call(context.Background(), "apply_patch", patchParams)
	if err != nil || patch.Hash.Hash == "" {
		t.Fatalf("patch result=%#v err=%v", patch, err)
	}
	run, err := registry.Call(context.Background(), "run_powershell", json.RawMessage(`{"command":"Get-Content -LiteralPath seed.txt"}`))
	if err != nil || run.Run.ExitCode != 0 || !strings.Contains(run.Run.Stdout, "needle new") {
		t.Fatalf("run result=%#v err=%v", run, err)
	}
	diff, err := registry.Call(context.Background(), "workspace_diff", json.RawMessage(`{"path":"seed.txt"}`))
	if err != nil || !strings.Contains(diff.Diff.Diff, "-needle old") || !strings.Contains(diff.Diff.Diff, "+needle new") {
		t.Fatalf("diff result=%#v err=%v", diff, err)
	}
}

func newTestRegistry(t *testing.T, root string, mode safety.Mode, approver safety.Approver, runner *brunelexec.Runner) *Registry {
	t.Helper()
	ws, err := workspace.Bind(root)
	if err != nil {
		t.Fatalf("bind workspace: %v", err)
	}
	return &Registry{
		Gate:      safety.NewGate(mode, approver, ws.Root()),
		Workspace: ws,
		Runner:    runner,
		ExecLimits: ExecLimits{
			DefaultTimeout: 10 * time.Second,
			MaxProcesses:   16,
			MaxMemoryBytes: 512 << 20,
			MaxOutputBytes: 1 << 20,
		},
	}
}

func assertFileContent(t *testing.T, path, want string) {
	t.Helper()
	actual, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	if string(actual) != want {
		t.Fatalf("content of %s = %q, want %q", path, actual, want)
	}
}

func gitRun(t *testing.T, root string, args ...string) {
	t.Helper()
	cmd := osexec.Command("git", append([]string{"-C", root}, args...)...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %s failed: %v\n%s", strings.Join(args, " "), err, output)
	}
}
