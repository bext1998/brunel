package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"math"
	"os"
	"strings"
	"time"

	brunelexec "github.com/bext1998/brunel/internal/exec"
	"github.com/bext1998/brunel/internal/safety"
	"github.com/bext1998/brunel/internal/tools"
	"github.com/bext1998/brunel/internal/workspace"
)

const (
	defaultMaxProcesses   = uint32(64)
	defaultMaxMemoryBytes = uint64(512 << 20)
	defaultMaxOutputBytes = int64(1 << 20)
)

var taylorToolNames = map[string]struct{}{
	"list_files":     {},
	"search_text":    {},
	"read_file":      {},
	"apply_patch":    {},
	"create_file":    {},
	"write_file":     {},
	"run_powershell": {},
	"workspace_diff": {},
}

type taylorToolConfig struct {
	name           string
	cwd            string
	mode           string
	timeout        time.Duration
	maxProcesses   uint
	maxMemoryBytes uint64
	maxOutputBytes int64
}

type taylorToolResponse struct {
	Tool      string        `json:"tool"`
	Result    *tools.Result `json:"result,omitempty"`
	Status    string        `json:"status"`
	ErrorCode string        `json:"error_code,omitempty"`
	Message   string        `json:"message,omitempty"`
}

func main() {
	os.Exit(run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}

func run(args []string, input io.Reader, output, diagnostics io.Writer) int {
	if !isTaylorToolInvocation(args) {
		_, _ = fmt.Fprintln(diagnostics, "brunel: not yet implemented (see #2)")
		return 1
	}

	config, err := parseTaylorToolConfig(args, diagnostics)
	if err != nil {
		writeTaylorToolResponse(output, taylorToolResponse{
			Tool:      config.name,
			Status:    "error",
			ErrorCode: tools.ErrInvalidArgument.Code,
			Message:   responseMessage(tools.ErrInvalidArgument.Code),
		})
		return 1
	}
	return runTaylorTool(context.Background(), config, input, output)
}

func isTaylorToolInvocation(args []string) bool {
	for _, arg := range args {
		if arg == "--taylor-tool" || strings.HasPrefix(arg, "--taylor-tool=") {
			return true
		}
	}
	return false
}

func parseTaylorToolConfig(args []string, diagnostics io.Writer) (taylorToolConfig, error) {
	config := taylorToolConfig{
		mode:           "workspace",
		timeout:        120 * time.Second,
		maxProcesses:   uint(defaultMaxProcesses),
		maxMemoryBytes: defaultMaxMemoryBytes,
		maxOutputBytes: defaultMaxOutputBytes,
	}
	flags := flag.NewFlagSet("brunel", flag.ContinueOnError)
	flags.SetOutput(diagnostics)
	flags.StringVar(&config.name, "taylor-tool", "", "run a frozen Taylor tool")
	flags.StringVar(&config.cwd, "cwd", "", "workspace directory")
	flags.StringVar(&config.mode, "mode", config.mode, "workspace or readonly")
	flags.DurationVar(&config.timeout, "timeout", config.timeout, "PowerShell timeout")
	flags.UintVar(&config.maxProcesses, "max-processes", config.maxProcesses, "PowerShell process limit")
	flags.Uint64Var(&config.maxMemoryBytes, "max-memory-bytes", config.maxMemoryBytes, "PowerShell per-process memory limit")
	flags.Int64Var(&config.maxOutputBytes, "max-output-bytes", config.maxOutputBytes, "PowerShell output limit per stream")
	if err := flags.Parse(args); err != nil {
		return config, err
	}
	if flags.NArg() != 0 || config.name == "" || !isTaylorToolName(config.name) {
		return config, tools.ErrInvalidArgument
	}
	if config.cwd == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return config, err
		}
		config.cwd = cwd
	}
	if config.mode != "workspace" && config.mode != "readonly" {
		return config, tools.ErrInvalidArgument
	}
	if config.timeout <= 0 || config.maxProcesses == 0 || config.maxProcesses > math.MaxUint32 || config.maxMemoryBytes == 0 || config.maxOutputBytes <= 0 {
		return config, tools.ErrInvalidArgument
	}
	return config, nil
}

func isTaylorToolName(name string) bool {
	_, ok := taylorToolNames[name]
	return ok
}

func runTaylorTool(ctx context.Context, config taylorToolConfig, input io.Reader, output io.Writer) int {
	params, err := io.ReadAll(input)
	if err != nil {
		return writeTaylorToolFailure(output, config.name, tools.ErrInvalidArgument.Code)
	}

	workspaceBinding, err := workspace.Bind(config.cwd)
	if err != nil {
		code := workspace.ErrorCode(err)
		if code == "" {
			code = workspace.ErrWorkspaceInvalid.Code
		}
		return writeTaylorToolFailure(output, config.name, code)
	}

	mode := safety.ModeWorkspace
	if config.mode == "readonly" {
		mode = safety.ModeReadonly
	}

	runner, runnerErr := brunelexec.NewRunner()
	registry := &tools.Registry{
		Workspace: workspaceBinding,
		Gate:      safety.NewGate(mode, nil, workspaceBinding.Root()),
		Runner:    runner,
		ExecLimits: tools.ExecLimits{
			DefaultTimeout: config.timeout,
			MaxProcesses:   uint32(config.maxProcesses),
			MaxMemoryBytes: config.maxMemoryBytes,
			MaxOutputBytes: config.maxOutputBytes,
		},
	}

	// This one-shot subprocess has no TTY or presentation layer. A CONFIRM
	// run_powershell call therefore fails closed with E_APPROVAL_REQUIRED_NO_TTY;
	// the approval UX belongs to issue #9's model-facing flow.
	result, err := registry.Call(ctx, config.name, json.RawMessage(params))
	if err != nil {
		if config.name == "run_powershell" && runner == nil && runnerErr != nil && tools.ErrorCode(err) == brunelexec.ErrUnsupportedPlatform.Code {
			err = runnerErr
		}
		code := tools.ErrorCode(err)
		if code == "" {
			code = tools.ErrToolIO.Code
		}
		return writeTaylorToolFailure(output, config.name, code)
	}

	writeTaylorToolResponse(output, taylorToolResponse{Tool: config.name, Result: &result, Status: "ok"})
	return 0
}

func writeTaylorToolFailure(output io.Writer, tool, code string) int {
	writeTaylorToolResponse(output, taylorToolResponse{
		Tool:      tool,
		Status:    "error",
		ErrorCode: code,
		Message:   responseMessage(code),
	})
	return 1
}

func writeTaylorToolResponse(output io.Writer, response taylorToolResponse) {
	if err := json.NewEncoder(output).Encode(response); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "brunel: write response:", err)
	}
}

func responseMessage(code string) string {
	switch code {
	case tools.ErrInvalidArgument.Code:
		return "invalid tool request"
	case workspace.ErrWorkspaceInvalid.Code:
		return "workspace is invalid"
	case safety.ErrReadonlyMode.Code:
		return "tool is unavailable in readonly mode"
	case safety.ErrApprovalRequiredNoTTY.Code:
		return "approval requires an interactive TTY"
	case brunelexec.ErrPwshRequired.Code:
		return "pwsh (PowerShell 7+) is required"
	case brunelexec.ErrUnsupportedPlatform.Code:
		return "operation is unsupported on this platform"
	default:
		return "tool call failed"
	}
}
