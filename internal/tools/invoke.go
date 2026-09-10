package tools

import (
	"context"
	"time"

	brunelexec "github.com/bext1998/brunel/internal/exec"
	"github.com/bext1998/brunel/internal/filetools"
	"github.com/bext1998/brunel/internal/safety"
)

func (r *Registry) decide(ctx context.Context, name, command string) error {
	if r.Gate == nil {
		return codeError(ErrInvalidArgument.Code, "safety gate is required", nil)
	}
	_, err := r.Gate.Decide(ctx, safety.ToolCall{Tool: name, Command: command})
	return err
}

func (r *Registry) workspaceRequired() error {
	if r.Workspace == nil {
		return codeError(ErrInvalidArgument.Code, "workspace is required", nil)
	}
	return nil
}

func (r *Registry) callListFiles(ctx context.Context, p ListFilesParams) (Result, error) {
	if err := r.decide(ctx, "list_files", ""); err != nil {
		return Result{}, err
	}
	if err := r.workspaceRequired(); err != nil {
		return Result{}, err
	}
	entries, err := listFiles(r.Workspace, p.Path, p.Glob, p.MaxDepth)
	if err != nil {
		return Result{}, err
	}
	return Result{Tool: "list_files", List: &ListResult{Entries: entries}}, nil
}

func (r *Registry) callSearchText(ctx context.Context, p SearchTextParams) (Result, error) {
	if err := r.decide(ctx, "search_text", ""); err != nil {
		return Result{}, err
	}
	if err := r.workspaceRequired(); err != nil {
		return Result{}, err
	}
	path := ""
	if p.Path != nil {
		path = *p.Path
	}
	matches, err := searchText(r.Workspace, p.Pattern, path, p.Glob, p.MaxResults)
	if err != nil {
		return Result{}, err
	}
	return Result{Tool: "search_text", Search: &SearchResult{Matches: matches}}, nil
}

func (r *Registry) callReadFile(ctx context.Context, p ReadFileParams) (Result, error) {
	if err := r.decide(ctx, "read_file", ""); err != nil {
		return Result{}, err
	}
	if err := r.workspaceRequired(); err != nil {
		return Result{}, err
	}
	start, end := 0, 0
	if p.StartLine != nil {
		start = *p.StartLine
	}
	if p.EndLine != nil {
		end = *p.EndLine
	}
	read, err := filetools.ReadFile(r.Workspace, p.Path, start, end)
	if err != nil {
		return Result{}, err
	}
	lines := make([]FileLine, len(read.Lines))
	for i, line := range read.Lines {
		lines[i] = FileLine{Number: line.Number, Text: line.Text}
	}
	return Result{Tool: "read_file", ReadFile: &ReadFileResult{Lines: lines, Hash: read.Hash}}, nil
}

func (r *Registry) callApplyPatch(ctx context.Context, p ApplyPatchParams) (Result, error) {
	if err := r.decide(ctx, "apply_patch", ""); err != nil {
		return Result{}, err
	}
	if err := r.workspaceRequired(); err != nil {
		return Result{}, err
	}
	hunks := make([]filetools.Hunk, len(p.Hunks))
	for i, hunk := range p.Hunks {
		hunks[i] = filetools.Hunk{StartLine: hunk.StartLine, EndLine: hunk.EndLine, OldLines: hunk.OldLines, NewLines: hunk.NewLines}
	}
	hash, err := filetools.ApplyPatch(r.Workspace, p.Path, p.ExpectedHash, hunks)
	if err != nil {
		return Result{}, err
	}
	return Result{Tool: "apply_patch", Hash: &HashResult{Hash: hash}}, nil
}

func (r *Registry) callCreateFile(ctx context.Context, p CreateFileParams) (Result, error) {
	if err := r.decide(ctx, "create_file", ""); err != nil {
		return Result{}, err
	}
	if err := r.workspaceRequired(); err != nil {
		return Result{}, err
	}
	hash, err := filetools.CreateFile(r.Workspace, p.Path, p.Content)
	if err != nil {
		return Result{}, err
	}
	return Result{Tool: "create_file", Hash: &HashResult{Hash: hash}}, nil
}

func (r *Registry) callWriteFile(ctx context.Context, p WriteFileParams) (Result, error) {
	if err := r.decide(ctx, "write_file", ""); err != nil {
		return Result{}, err
	}
	if err := r.workspaceRequired(); err != nil {
		return Result{}, err
	}
	hash, err := filetools.WriteFile(r.Workspace, p.Path, p.ExpectedHash, p.Content)
	if err != nil {
		return Result{}, err
	}
	return Result{Tool: "write_file", Hash: &HashResult{Hash: hash}}, nil
}

func (r *Registry) callRunPowerShell(ctx context.Context, p RunPowerShellParams) (Result, error) {
	if err := r.decide(ctx, "run_powershell", p.Command); err != nil {
		return Result{}, err
	}
	if err := r.workspaceRequired(); err != nil {
		return Result{}, err
	}
	if err := validateExecLimits(r.ExecLimits); err != nil {
		return Result{}, err
	}
	cwdPath := ""
	if p.CWD != nil {
		cwdPath = *p.CWD
	}
	cwd, err := r.Workspace.Resolve(cwdPath)
	if err != nil {
		return Result{}, err
	}
	if r.Runner == nil {
		return Result{}, brunelexec.ErrUnsupportedPlatform
	}
	timeout := r.ExecLimits.DefaultTimeout
	if p.TimeoutSec != nil {
		timeout = time.Duration(*p.TimeoutSec) * time.Second
	}
	output, err := r.Runner.Run(ctx, brunelexec.Options{
		Command:        p.Command,
		WorkDir:        cwd,
		Timeout:        timeout,
		MaxProcesses:   r.ExecLimits.MaxProcesses,
		MaxMemoryBytes: r.ExecLimits.MaxMemoryBytes,
		MaxOutputBytes: r.ExecLimits.MaxOutputBytes,
	})
	if err != nil {
		return Result{}, err
	}
	return Result{Tool: "run_powershell", Run: &RunResult{
		Stdout: string(output.Stdout), Stderr: string(output.Stderr), ExitCode: output.ExitCode,
		Truncated: output.StdoutTruncated || output.StderrTruncated,
	}}, nil
}

func validateExecLimits(limits ExecLimits) error {
	if limits.DefaultTimeout <= 0 || limits.MaxProcesses == 0 || limits.MaxMemoryBytes == 0 || limits.MaxOutputBytes <= 0 {
		return codeError(ErrInvalidArgument.Code, "all PowerShell execution limits must be non-zero", nil)
	}
	return nil
}

func (r *Registry) callWorkspaceDiff(ctx context.Context, p WorkspaceDiffParams) (Result, error) {
	if err := r.decide(ctx, "workspace_diff", ""); err != nil {
		return Result{}, err
	}
	if err := r.workspaceRequired(); err != nil {
		return Result{}, err
	}
	path := ""
	if p.Path != nil {
		path = *p.Path
	}
	diff, err := workspaceDiff(ctx, r.Workspace, r.Workspace.Root(), path)
	if err != nil {
		return Result{}, err
	}
	return Result{Tool: "workspace_diff", Diff: &DiffResult{Diff: diff}}, nil
}
