package tools

import (
	"context"
	"encoding/json"
	"time"

	brunelexec "github.com/bext1998/brunel/internal/exec"
	"github.com/bext1998/brunel/internal/safety"
	"github.com/bext1998/brunel/internal/workspace"
)

// ExecLimits supplies the explicit resource limits required by exec.Options.
// Policy belongs to the caller; tools never invents a limit value.
type ExecLimits struct {
	DefaultTimeout time.Duration
	MaxProcesses   uint32
	MaxMemoryBytes uint64
	MaxOutputBytes int64
}

// Registry wires the frozen tools to Brunel's workspace, safety gate, and
// PowerShell runner. Workspace also satisfies filetools.Resolver.
type Registry struct {
	Gate       *safety.Gate
	Workspace  *workspace.Workspace
	Runner     *brunelexec.Runner
	ExecLimits ExecLimits
}

// Call strictly validates one frozen tool call then executes it. Decide runs
// before any filesystem, Git, or PowerShell access (INV-1); path resolution
// and stale-hash checks are preconditions that occur only after Decide.
func (r *Registry) Call(ctx context.Context, name string, params json.RawMessage) (Result, error) {
	if r == nil {
		return Result{}, codeError(ErrInvalidArgument.Code, "tools registry is nil", nil)
	}
	switch name {
	case "list_files":
		var p ListFilesParams
		if err := decodeParams(params, &p, "path"); err != nil {
			return Result{}, err
		}
		if err := validateListFilesParams(p); err != nil {
			return Result{}, err
		}
		return r.callListFiles(ctx, p)
	case "search_text":
		var p SearchTextParams
		if err := decodeParams(params, &p, "pattern"); err != nil {
			return Result{}, err
		}
		if err := validateSearchTextParams(p); err != nil {
			return Result{}, err
		}
		return r.callSearchText(ctx, p)
	case "read_file":
		var p ReadFileParams
		if err := decodeParams(params, &p, "path"); err != nil {
			return Result{}, err
		}
		if err := validateReadFileParams(p); err != nil {
			return Result{}, err
		}
		return r.callReadFile(ctx, p)
	case "apply_patch":
		var p ApplyPatchParams
		if err := decodeApplyPatchParams(params, &p); err != nil {
			return Result{}, err
		}
		if err := validateApplyPatchParams(p); err != nil {
			return Result{}, err
		}
		return r.callApplyPatch(ctx, p)
	case "create_file":
		var p CreateFileParams
		if err := decodeParams(params, &p, "path", "content"); err != nil {
			return Result{}, err
		}
		if err := validateCreateFileParams(p); err != nil {
			return Result{}, err
		}
		return r.callCreateFile(ctx, p)
	case "write_file":
		var p WriteFileParams
		if err := decodeParams(params, &p, "path", "expected_hash", "content"); err != nil {
			return Result{}, err
		}
		if err := validateWriteFileParams(p); err != nil {
			return Result{}, err
		}
		return r.callWriteFile(ctx, p)
	case "run_powershell":
		var p RunPowerShellParams
		if err := decodeParams(params, &p, "command"); err != nil {
			return Result{}, err
		}
		if err := validateRunPowerShellParams(p); err != nil {
			return Result{}, err
		}
		return r.callRunPowerShell(ctx, p)
	case "workspace_diff":
		var p WorkspaceDiffParams
		if err := decodeParams(params, &p); err != nil {
			return Result{}, err
		}
		return r.callWorkspaceDiff(ctx, p)
	default:
		return Result{}, codeError(ErrInvalidArgument.Code, "unknown tool "+name, nil)
	}
}
