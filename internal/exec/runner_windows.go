//go:build windows

package exec

import (
	"context"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	osexec "os/exec"
)

// Runner executes a single PowerShell 7 command per Run call, bound into
// a fresh Windows Job Object so the whole process tree it spawns is
// deterministically terminated on timeout, cancellation, or completion.
type Runner struct {
	pwshPath string
	mu       sync.Mutex
}

// NewRunner locates pwsh (PowerShell 7+) on PATH. It never falls back to
// Windows PowerShell 5.1 (powershell.exe) or cmd.exe.
func NewRunner() (*Runner, error) {
	path, err := osexec.LookPath("pwsh")
	if err != nil {
		return nil, codeError(ErrPwshRequired.Code, "pwsh (PowerShell 7+) was not found on PATH", err)
	}
	return &Runner{pwshPath: path}, nil
}

func validateOptions(opts Options) error {
	if strings.TrimSpace(opts.Command) == "" {
		return codeError(ErrInvalidArgument.Code, "command is empty", nil)
	}
	if opts.WorkDir != "" && !filepath.IsAbs(opts.WorkDir) {
		return codeError(ErrInvalidArgument.Code, "work dir must be an absolute path", nil)
	}
	if opts.Timeout <= 0 {
		return codeError(ErrInvalidArgument.Code, "timeout must be > 0", nil)
	}
	if opts.MaxProcesses == 0 {
		return codeError(ErrInvalidArgument.Code, "max processes must be > 0", nil)
	}
	if opts.MaxMemoryBytes == 0 {
		return codeError(ErrInvalidArgument.Code, "max memory bytes must be > 0", nil)
	}
	if opts.MaxOutputBytes <= 0 {
		return codeError(ErrInvalidArgument.Code, "max output bytes must be > 0", nil)
	}
	return nil
}

// testHookAfterAssignBeforeResume, when non-nil, is invoked after the
// child process has been assigned to the job object and before its
// suspended thread is resumed. Only ever set by white-box tests in this
// package; production code paths leave it nil.
var testHookAfterAssignBeforeResume func(job, process syscall.Handle)

func (r *Runner) Run(ctx context.Context, opts Options) (Output, error) {
	if err := validateOptions(opts); err != nil {
		return Output{}, err
	}

	// One command at a time per Runner: keeps the set of inheritable
	// handles alive at CreateProcess time limited to the ones this call
	// itself creates (see plan §"Handle 繼承的安全取捨").
	r.mu.Lock()
	defer r.mu.Unlock()

	cmdLine := buildCommandLine(r.pwshPath, opts.Command)

	job, err := createJobObject()
	if err != nil {
		return Output{}, codeError(ErrExecInternal.Code, "cannot create job object", err)
	}
	defer syscall.CloseHandle(job) // KILL_ON_JOB_CLOSE backstop for any path below that returns early

	if err := setJobLimits(job, opts.MaxProcesses, opts.MaxMemoryBytes); err != nil {
		return Output{}, codeError(ErrExecInternal.Code, "cannot set job object limits", err)
	}

	stdin, err := openNulReadHandle()
	if err != nil {
		return Output{}, codeError(ErrExecInternal.Code, "cannot open NUL for stdin", err)
	}
	defer syscall.CloseHandle(stdin)

	stdoutRead, stdoutWrite, err := createOutputPipe()
	if err != nil {
		return Output{}, codeError(ErrExecInternal.Code, "cannot create stdout pipe", err)
	}
	defer syscall.CloseHandle(stdoutRead)

	stderrRead, stderrWrite, err := createOutputPipe()
	if err != nil {
		syscall.CloseHandle(stdoutWrite)
		return Output{}, codeError(ErrExecInternal.Code, "cannot create stderr pipe", err)
	}
	defer syscall.CloseHandle(stderrRead)

	pi, err := startSuspendedProcess(opts.WorkDir, cmdLine, stdin, stdoutWrite, stderrWrite)
	// The child inherited its own copies of the write ends; the parent's
	// copies must close now regardless of outcome, or the read side will
	// never see EOF once the child exits.
	syscall.CloseHandle(stdoutWrite)
	syscall.CloseHandle(stderrWrite)
	if err != nil {
		return Output{}, codeError(ErrExecInternal.Code, "cannot start pwsh", err)
	}
	defer syscall.CloseHandle(pi.Process)

	if err := assignProcessToJobObject(job, pi.Process); err != nil {
		// Not yet job-bound: this is the only path where a direct
		// TerminateProcess (rather than TerminateJobObject) is correct.
		syscall.CloseHandle(pi.Thread)
		syscall.TerminateProcess(pi.Process, 1)
		return Output{}, codeError(ErrExecInternal.Code, "cannot assign process to job object", err)
	}

	if testHookAfterAssignBeforeResume != nil {
		testHookAfterAssignBeforeResume(job, pi.Process)
	}

	if _, err := resumeThread(pi.Thread); err != nil {
		syscall.CloseHandle(pi.Thread)
		terminateJobObject(job, 1)
		return Output{}, codeError(ErrExecInternal.Code, "cannot resume suspended process", err)
	}
	syscall.CloseHandle(pi.Thread)

	stdoutCh := drainPipe(stdoutRead, opts.MaxOutputBytes)
	stderrCh := drainPipe(stderrRead, opts.MaxOutputBytes)

	exitCh := make(chan struct{})
	go func() {
		syscall.WaitForSingleObject(pi.Process, syscall.INFINITE)
		close(exitCh)
	}()

	timer := time.NewTimer(opts.Timeout)
	defer timer.Stop()

	var runErr error
	select {
	case <-exitCh:
	case <-timer.C:
		runErr = ErrToolTimeout
	case <-ctx.Done():
		runErr = ctx.Err()
	}

	var exitCode uint32
	if runErr == nil {
		_ = syscall.GetExitCodeProcess(pi.Process, &exitCode) // best effort; falls back to -1 below on failure
	}

	// Unconditional, even on the happy path: guarantees no descendant of
	// this command outlives Run(), matching "整棵程序樹終止" for every
	// exit path, not just timeout/cancel. Safe/idempotent once the
	// primary process has already exited on its own.
	terminateJobObject(job, 1)
	<-exitCh // exitCh is closed, not just signaled once, so this never blocks here

	stdout := <-stdoutCh
	stderr := <-stderrCh

	out := Output{
		Stdout:          stdout.data,
		StdoutTruncated: stdout.truncated,
		Stderr:          stderr.data,
		StderrTruncated: stderr.truncated,
		ExitCode:        -1,
	}
	if runErr == nil {
		out.ExitCode = int(exitCode)
	}
	return out, runErr
}

type pipeResult struct {
	data      []byte
	truncated bool
}

// drainPipe reads h until EOF (i.e. every inheritable copy of its write
// end has closed), capping the retained bytes at maxBytes. Once the cap
// is hit it keeps reading and discarding rather than stopping, so a
// chatty child never blocks on a full pipe buffer.
func drainPipe(h syscall.Handle, maxBytes int64) <-chan pipeResult {
	ch := make(chan pipeResult, 1)
	go func() {
		defer syscall.CloseHandle(h)
		buf := make([]byte, 32*1024)
		var out []byte
		truncated := false
		for {
			n, err := syscall.Read(h, buf)
			if n > 0 {
				if !truncated {
					remaining := maxBytes - int64(len(out))
					if remaining > 0 {
						take := int64(n)
						if take > remaining {
							take = remaining
						}
						out = append(out, buf[:take]...)
					}
					if int64(len(out)) >= maxBytes {
						truncated = true
					}
				}
			}
			// syscall.Read maps ERROR_BROKEN_PIPE (the write end has
			// closed) to (0, nil) rather than an error, so n == 0 is
			// the real EOF signal here, not just "no data yet" -- a
			// blocking synchronous pipe read never returns (0, nil)
			// for any other reason.
			if n == 0 || err != nil {
				break
			}
		}
		ch <- pipeResult{data: out, truncated: truncated}
	}()
	return ch
}
