//go:build windows

package exec

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"
)

func newTestRunner(t *testing.T) *Runner {
	t.Helper()
	r, err := NewRunner()
	if err != nil {
		t.Fatalf("NewRunner() error = %v", err)
	}
	return r
}

func baseOptions(command string) Options {
	return Options{
		Command:        command,
		Timeout:        10 * time.Second,
		MaxProcesses:   16,
		MaxMemoryBytes: 512 * 1024 * 1024,
		MaxOutputBytes: 64 * 1024,
	}
}

func TestPSRunner_HappyPath(t *testing.T) {
	r := newTestRunner(t)
	out, err := r.Run(context.Background(), baseOptions("Write-Output 'hello'; exit 3"))
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if out.ExitCode != 3 {
		t.Fatalf("ExitCode = %d, want 3", out.ExitCode)
	}
	if strings.TrimSpace(string(out.Stdout)) != "hello" {
		t.Fatalf("Stdout = %q, want %q", out.Stdout, "hello")
	}
	if out.StdoutTruncated || out.StderrTruncated {
		t.Fatalf("unexpected truncation: stdout=%v stderr=%v", out.StdoutTruncated, out.StderrTruncated)
	}
}

func TestPSRunner_InvalidOptions(t *testing.T) {
	r := newTestRunner(t)
	base := baseOptions("Write-Output 'x'")

	cases := []struct {
		name string
		mut  func(*Options)
	}{
		{"empty command", func(o *Options) { o.Command = "   " }},
		{"zero timeout", func(o *Options) { o.Timeout = 0 }},
		{"zero max processes", func(o *Options) { o.MaxProcesses = 0 }},
		{"zero max memory", func(o *Options) { o.MaxMemoryBytes = 0 }},
		{"zero max output", func(o *Options) { o.MaxOutputBytes = 0 }},
		{"relative work dir", func(o *Options) { o.WorkDir = `relative\path` }},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			opts := base
			c.mut(&opts)
			if _, err := r.Run(context.Background(), opts); !errors.Is(err, ErrInvalidArgument) {
				t.Fatalf("Run() error = %v, want E_INVALID_ARGUMENT", err)
			}
		})
	}
}

func TestPSRunner_OutputTruncated(t *testing.T) {
	r := newTestRunner(t)
	opts := baseOptions(`1..20000 | ForEach-Object { Write-Output 'line of output padding to grow the buffer past the cap' }`)
	opts.MaxOutputBytes = 4096

	out, err := r.Run(context.Background(), opts)
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if !out.StdoutTruncated {
		t.Fatalf("expected StdoutTruncated = true")
	}
	if int64(len(out.Stdout)) != opts.MaxOutputBytes {
		t.Fatalf("len(Stdout) = %d, want %d", len(out.Stdout), opts.MaxOutputBytes)
	}
	if out.ExitCode != 0 {
		t.Fatalf("ExitCode = %d, want 0 (truncation must not escalate to failure)", out.ExitCode)
	}
}

// TestPSRunner_Timeout_KillsProcessTree covers AC-10 / TC-EXEC: a
// grandchild process must die along with its parent when the tool call
// times out, not just the directly-spawned pwsh.
func TestPSRunner_Timeout_KillsProcessTree(t *testing.T) {
	r := newTestRunner(t)
	counterPath := filepath.Join(t.TempDir(), "counter.txt")
	t.Setenv("BRUNEL_TEST_COUNTER_FILE", counterPath)

	script := `
$psi = New-Object System.Diagnostics.ProcessStartInfo
$psi.FileName = 'pwsh'
$psi.ArgumentList.Add('-NoProfile')
$psi.ArgumentList.Add('-NonInteractive')
$psi.ArgumentList.Add('-Command')
$psi.ArgumentList.Add('$i=0; while ($true) { Set-Content -Path $env:BRUNEL_TEST_COUNTER_FILE -Value $i; $i++; Start-Sleep -Milliseconds 100 }')
$psi.UseShellExecute = $false
[System.Diagnostics.Process]::Start($psi) | Out-Null
Start-Sleep -Seconds 30
`
	opts := baseOptions(script)
	opts.Timeout = 2 * time.Second

	start := time.Now()
	_, err := r.Run(context.Background(), opts)
	elapsed := time.Since(start)

	if !errors.Is(err, ErrToolTimeout) {
		t.Fatalf("Run() error = %v, want E_TOOL_TIMEOUT", err)
	}
	if elapsed > 10*time.Second {
		t.Fatalf("Run() took %v, expected to return shortly after the %v timeout", elapsed, opts.Timeout)
	}

	v0, ok := readCounter(t, counterPath)
	if !ok {
		t.Fatalf("grandchild never wrote a counter value; process tree may not have started")
	}
	time.Sleep(700 * time.Millisecond)
	v1, _ := readCounter(t, counterPath)
	if v1 != v0 {
		t.Fatalf("counter still advancing after timeout (v0=%d v1=%d): grandchild survived Job Object termination", v0, v1)
	}
}

func readCounter(t *testing.T, path string) (int, bool) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		return 0, false
	}
	v, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil {
		return 0, false
	}
	return v, true
}

// TestPSRunner_JobBoundBeforeResume covers the escape-window requirement
// from spec §10.3: AssignProcessToJobObject must have already completed
// before the suspended thread ever runs a single instruction.
func TestPSRunner_JobBoundBeforeResume(t *testing.T) {
	r := newTestRunner(t)
	markerPath := filepath.Join(t.TempDir(), "marker.txt")
	t.Setenv("BRUNEL_TEST_MARKER_FILE", markerPath)

	var mu sync.Mutex
	var hookErr error
	var sawMember bool

	testHookAfterAssignBeforeResume = func(job, process syscall.Handle) {
		mu.Lock()
		defer mu.Unlock()
		// Independent evidence the thread truly hasn't run yet: its
		// first scripted action (writing the marker file) must not
		// have happened at the point this hook fires.
		if _, statErr := os.Stat(markerPath); !os.IsNotExist(statErr) {
			hookErr = errors.New("marker file already exists before ResumeThread; child ran before job assignment")
			return
		}
		member, err := isProcessInJob(process, job)
		if err != nil {
			hookErr = err
			return
		}
		sawMember = member
	}
	defer func() { testHookAfterAssignBeforeResume = nil }()

	_, err := r.Run(context.Background(), baseOptions(`Set-Content -Path $env:BRUNEL_TEST_MARKER_FILE -Value '1'`))
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	mu.Lock()
	defer mu.Unlock()
	if hookErr != nil {
		t.Fatalf("hook observed a problem: %v", hookErr)
	}
	if !sawMember {
		t.Fatalf("process was not a job member at the point AssignProcessToJobObject returned, before ResumeThread")
	}
	if _, err := os.Stat(markerPath); err != nil {
		t.Fatalf("marker file was not written after Run() returned: %v", err)
	}
}
