package exec

import "time"

// Options configures a single Run call. All limit fields are required
// (must be non-zero); this package never applies a default, since the
// real policy values are decided by callers, not by exec.
type Options struct {
	Command        string        // passed verbatim as pwsh's single -Command argument
	WorkDir        string        // absolute path; caller is responsible for escape validation
	Timeout        time.Duration // required, must be > 0
	MaxProcesses   uint32        // required, must be > 0 (Job Object ActiveProcessLimit)
	MaxMemoryBytes uint64        // required, must be > 0 (Job Object ProcessMemoryLimit, per process)
	MaxOutputBytes int64         // required, must be > 0 (stdout/stderr capture cap, each)
}

// Output is what a Run call captured, regardless of whether it returned
// an error. On timeout or context cancellation, Output still reflects
// whatever was captured before the process tree was terminated.
type Output struct {
	Stdout          []byte
	Stderr          []byte
	ExitCode        int // only meaningful when Run returned a nil error
	StdoutTruncated bool
	StderrTruncated bool
}
