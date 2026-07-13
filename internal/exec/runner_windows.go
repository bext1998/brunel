//go:build windows

package exec

import (
	"context"
	"errors"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	osexec "os/exec"
)

// Runner executes PowerShell 7 commands. Each Run call is bound into its
// own fresh Windows Job Object so the whole process tree it spawns is
// deterministically terminated on timeout, cancellation, or completion.
// A Runner may be used for multiple concurrent Run calls; see
// createProcessMu for the one part of that path that must still be
// serialized process-wide.
type Runner struct {
	pwshPath string
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

// createProcessMu serializes every Runner's handle-creation-through-
// CreateProcess span, process-wide, not just per Runner. CreateProcess
// with bInheritHandles=true duplicates *every* currently-inheritable
// handle in the process into the new child, not just the ones referenced
// in STARTUPINFO -- so two concurrent Run calls, even on separate Runner
// instances, could otherwise have one child inherit another's still-open
// pipe write end, which would prevent that pipe from ever seeing EOF.
var createProcessMu sync.Mutex

// terminationGracePeriod bounds how long Run waits to confirm the
// process tree actually died after a Job Object (or, as a fallback,
// direct process) termination request. It exists so a failed
// termination call can never hang Run forever.
const terminationGracePeriod = 5 * time.Second

func (r *Runner) Run(ctx context.Context, opts Options) (Output, error) {
	if err := validateOptions(opts); err != nil {
		return Output{}, err
	}

	cmdLine := buildCommandLine(r.pwshPath, opts.Command)

	job, err := createJobObject()
	if err != nil {
		return Output{}, codeError(ErrExecInternal.Code, "cannot create job object", err)
	}
	defer syscall.CloseHandle(job) // KILL_ON_JOB_CLOSE backstop for any path below that returns early

	if err := setJobLimits(job, opts.MaxProcesses, opts.MaxMemoryBytes); err != nil {
		return Output{}, codeError(ErrExecInternal.Code, "cannot set job object limits", err)
	}

	createProcessMu.Lock()

	stdin, err := openNulReadHandle()
	if err != nil {
		createProcessMu.Unlock()
		return Output{}, codeError(ErrExecInternal.Code, "cannot open NUL for stdin", err)
	}

	stdoutRead, stdoutWrite, err := createOutputPipe()
	if err != nil {
		syscall.CloseHandle(stdin)
		createProcessMu.Unlock()
		return Output{}, codeError(ErrExecInternal.Code, "cannot create stdout pipe", err)
	}
	// stdoutRead's ownership transfers to drainPipe below once Run
	// actually reaches it; until then, Run itself must close it on every
	// early-return path (avoids the double-close that would otherwise
	// happen if both this defer and drainPipe's own close ran).
	closeStdoutRead := true
	defer func() {
		if closeStdoutRead {
			syscall.CloseHandle(stdoutRead)
		}
	}()

	stderrRead, stderrWrite, err := createOutputPipe()
	if err != nil {
		syscall.CloseHandle(stdin)
		syscall.CloseHandle(stdoutWrite)
		createProcessMu.Unlock()
		return Output{}, codeError(ErrExecInternal.Code, "cannot create stderr pipe", err)
	}
	closeStderrRead := true
	defer func() {
		if closeStderrRead {
			syscall.CloseHandle(stderrRead)
		}
	}()

	pi, err := startSuspendedProcess(opts.WorkDir, cmdLine, stdin, stdoutWrite, stderrWrite)
	// The child inherited its own copies of these; the parent's copies
	// must close now regardless of outcome -- both so the read sides can
	// see EOF once the child exits, and to end the process-wide
	// inheritance window createProcessMu protects.
	syscall.CloseHandle(stdin)
	syscall.CloseHandle(stdoutWrite)
	syscall.CloseHandle(stderrWrite)
	createProcessMu.Unlock()
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

	// From here on, drainPipe owns stdoutRead/stderrRead and closes them
	// itself; Run must not close them again on the way out.
	closeStdoutRead = false
	closeStderrRead = false
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
	termErr := terminateJobObject(job, 1)
	if termErr != nil {
		// Job-level kill failed; fall back to killing the tracked
		// process directly so there is still a path to a confirmed exit.
		// If this ALSO fails, join it into termErr rather than
		// discarding it -- both are relevant if the bounded wait below
		// ultimately gives up.
		if tpErr := syscall.TerminateProcess(pi.Process, 1); tpErr != nil {
			termErr = errors.Join(termErr, tpErr)
		}
	}

	stdout, stderr, waitErr := waitForExitAndOutput(exitCh, stdoutCh, stderrCh, terminationGracePeriod, termErr)
	if waitErr != nil {
		return Output{}, waitErr
	}

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

// waitForExitAndOutput waits for confirmed process exit and both pipes
// to reach EOF, all against one shared deadline. A single deadline
// (rather than one per wait) matters: if Job Object termination fails
// and the process-level fallback only kills the directly-tracked
// process, a grandchild outside the job's reach could still hold a
// pipe's write end open indefinitely -- so the pipe drains must be
// bounded by the same grace period as the exit-confirmation wait, not
// left as unconditional receives, or that scenario still hangs Run
// forever despite the exit-side bound.
func waitForExitAndOutput(exitCh <-chan struct{}, stdoutCh, stderrCh <-chan pipeResult, deadline time.Duration, cause error) (pipeResult, pipeResult, error) {
	timer := time.NewTimer(deadline)
	defer timer.Stop()

	select {
	case <-exitCh: // closed, not just signaled once, so an earlier fire here is fine
	case <-timer.C:
		// Could not confirm the process tree actually died. Returning
		// here instead of blocking forever preserves the deterministic-
		// termination contract, at the cost of admitting termination
		// could not be verified.
		return pipeResult{}, pipeResult{}, codeError(ErrExecInternal.Code, "could not confirm process tree termination within the grace period", cause)
	}

	var stdout, stderr pipeResult
	select {
	case stdout = <-stdoutCh:
	case <-timer.C:
		return pipeResult{}, pipeResult{}, codeError(ErrExecInternal.Code, "stdout pipe did not reach EOF after process tree termination (a descendant may still hold it open)", cause)
	}
	select {
	case stderr = <-stderrCh:
	case <-timer.C:
		return pipeResult{}, pipeResult{}, codeError(ErrExecInternal.Code, "stderr pipe did not reach EOF after process tree termination (a descendant may still hold it open)", cause)
	}
	return stdout, stderr, nil
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
