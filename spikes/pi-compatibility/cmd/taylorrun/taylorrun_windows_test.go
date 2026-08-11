//go:build windows

package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	brunellexec "github.com/bext1998/brunel/internal/exec"
)

func TestPiSurrogateHelper(t *testing.T) {
	if os.Getenv("BRUNEL_PI_SURROGATE") != "1" {
		return
	}

	heartbeatPath := os.Getenv("BRUNEL_PI_HEARTBEAT")
	if heartbeatPath == "" {
		os.Exit(2)
	}
	for i := 0; ; i++ {
		if err := os.WriteFile(heartbeatPath, []byte(strconv.Itoa(i)), 0o600); err != nil {
			os.Exit(3)
		}
		time.Sleep(50 * time.Millisecond)
	}
}

func startPiSurrogate(t *testing.T) (*os.Process, string) {
	t.Helper()
	heartbeatPath := filepath.Join(t.TempDir(), "pi-heartbeat.txt")
	cmd := exec.Command(os.Args[0], "-test.run=^TestPiSurrogateHelper$")
	cmd.Env = append(os.Environ(),
		"BRUNEL_PI_SURROGATE=1",
		"BRUNEL_PI_HEARTBEAT="+heartbeatPath,
	)
	if err := cmd.Start(); err != nil {
		t.Fatalf("start Pi surrogate: %v", err)
	}
	t.Cleanup(func() {
		_ = cmd.Process.Kill()
		_, _ = cmd.Process.Wait()
	})
	if !waitForFile(t, heartbeatPath, 5*time.Second) {
		t.Fatalf("Pi surrogate did not publish a heartbeat")
	}
	return cmd.Process, heartbeatPath
}

func TestPiAbortDoesNotOwnGoJob(t *testing.T) {
	_, heartbeatPath := startPiSurrogate(t)
	runner := newSpikeRunner(t)
	counterPath := filepath.Join(t.TempDir(), "counter.txt")
	command := longRunningChildCommand(counterPath)

	ctx, cancel := context.WithCancel(context.Background())
	resultCh := make(chan error, 1)
	go func() {
		_, err := runner.Run(ctx, spikeOptions(command))
		resultCh <- err
	}()

	if !waitForFile(t, counterPath, 5*time.Second) {
		t.Fatalf("Job Object child did not start")
	}
	cancel()

	select {
	case err := <-resultCh:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("Run() error = %v, want context.Canceled", err)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("Run() did not return after Pi abort simulation")
	}

	assertCounterStopped(t, counterPath)
	assertHeartbeatAdvancing(t, heartbeatPath)
}

func TestPiCrashDoesNotOwnGoJob(t *testing.T) {
	piProcess, heartbeatPath := startPiSurrogate(t)
	runner := newSpikeRunner(t)
	counterPath := filepath.Join(t.TempDir(), "counter.txt")
	command := longRunningChildCommand(counterPath)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	resultCh := make(chan error, 1)
	go func() {
		_, err := runner.Run(ctx, spikeOptions(command))
		resultCh <- err
	}()

	if !waitForFile(t, counterPath, 5*time.Second) {
		t.Fatalf("Job Object child did not start")
	}
	if err := piProcess.Kill(); err != nil {
		t.Fatalf("kill Pi surrogate: %v", err)
	}
	if _, err := piProcess.Wait(); err != nil {
		// Process.Kill normally makes Wait return a non-nil exit status; the
		// important evidence is that the process has stopped.
		t.Logf("Pi surrogate Wait after simulated crash: %v", err)
	}
	assertHeartbeatStopped(t, heartbeatPath)

	before := readCounter(t, counterPath)
	time.Sleep(500 * time.Millisecond)
	after := readCounter(t, counterPath)
	if after <= before {
		t.Fatalf("Go Job Object child stopped after Pi crash (before=%d after=%d)", before, after)
	}

	cancel()
	select {
	case err := <-resultCh:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("Run() cleanup error = %v, want context.Canceled", err)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("Run() did not return after test cleanup")
	}
	assertCounterStopped(t, counterPath)
}

func newSpikeRunner(t *testing.T) *brunellexec.Runner {
	t.Helper()
	runner, err := brunellexec.NewRunner()
	if err != nil {
		t.Fatalf("NewRunner(): %v", err)
	}
	return runner
}

func spikeOptions(command string) brunellexec.Options {
	return brunellexec.Options{
		Command:        command,
		Timeout:        20 * time.Second,
		MaxProcesses:   16,
		MaxMemoryBytes: 512 * 1024 * 1024,
		MaxOutputBytes: 64 * 1024,
	}
}

func longRunningChildCommand(counterPath string) string {
	quotedPath := strings.ReplaceAll(counterPath, "'", "''")
	return fmt.Sprintf(`
$psi = New-Object System.Diagnostics.ProcessStartInfo
$psi.FileName = 'pwsh'
$psi.ArgumentList.Add('-NoProfile')
$psi.ArgumentList.Add('-NonInteractive')
$psi.ArgumentList.Add('-Command')
$psi.ArgumentList.Add('$i=0; while ($true) { Set-Content -Path ''%s'' -Value $i; $i++; Start-Sleep -Milliseconds 50 }')
$psi.UseShellExecute = $false
[System.Diagnostics.Process]::Start($psi) | Out-Null
Start-Sleep -Seconds 30
`, quotedPath)
}

func waitForFile(t *testing.T, path string, timeout time.Duration) bool {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if _, err := os.Stat(path); err == nil {
			return true
		}
		time.Sleep(50 * time.Millisecond)
	}
	return false
}

func readCounter(t *testing.T, path string) int {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		data, err := os.ReadFile(path)
		if err == nil {
			if value, parseErr := strconv.Atoi(strings.TrimSpace(string(data))); parseErr == nil {
				return value
			}
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatalf("could not read numeric counter from %s", path)
	return 0
}

func assertCounterStopped(t *testing.T, path string) {
	t.Helper()
	before := readCounter(t, path)
	time.Sleep(500 * time.Millisecond)
	after := readCounter(t, path)
	if after != before {
		t.Fatalf("Job Object descendant still alive after cleanup (before=%d after=%d)", before, after)
	}
}

func assertHeartbeatAdvancing(t *testing.T, path string) {
	t.Helper()
	first := readCounter(t, path)
	time.Sleep(250 * time.Millisecond)
	second := readCounter(t, path)
	if second <= first {
		t.Fatalf("Pi surrogate heartbeat stopped (first=%d second=%d)", first, second)
	}
}

func assertHeartbeatStopped(t *testing.T, path string) {
	t.Helper()
	first := readCounter(t, path)
	time.Sleep(250 * time.Millisecond)
	second := readCounter(t, path)
	if second != first {
		t.Fatalf("Pi surrogate heartbeat continued after simulated crash (first=%d second=%d)", first, second)
	}
}
