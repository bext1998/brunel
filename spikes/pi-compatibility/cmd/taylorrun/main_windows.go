//go:build windows

package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	brunellexec "github.com/bext1998/brunel/internal/exec"
)

type response struct {
	Stdout          string `json:"stdout"`
	Stderr          string `json:"stderr"`
	ExitCode        int    `json:"exitCode"`
	StdoutTruncated bool   `json:"stdoutTruncated"`
	StderrTruncated bool   `json:"stderrTruncated"`
	ErrorCode       string `json:"errorCode,omitempty"`
	Error           string `json:"error,omitempty"`
}

func main() {
	var (
		command        string
		workDir        string
		timeout        time.Duration
		maxProcesses   uint
		maxMemoryBytes uint64
		maxOutputBytes int64
	)
	flag.StringVar(&command, "command", "", "PowerShell command to execute")
	flag.StringVar(&workDir, "workdir", "", "absolute working directory (defaults to the current directory)")
	flag.DurationVar(&timeout, "timeout", 30*time.Second, "execution timeout")
	flag.UintVar(&maxProcesses, "max-processes", 32, "Job Object active process limit")
	flag.Uint64Var(&maxMemoryBytes, "max-memory-bytes", 512*1024*1024, "per-process memory limit")
	flag.Int64Var(&maxOutputBytes, "max-output-bytes", 64*1024, "stdout/stderr retention cap")
	flag.Parse()

	if workDir == "" {
		var err error
		workDir, err = os.Getwd()
		if err != nil {
			writeResponse(response{ErrorCode: "E_WORKDIR", Error: err.Error()})
			os.Exit(1)
		}
	}
	if absolute, err := filepath.Abs(workDir); err == nil {
		workDir = absolute
	}

	runner, err := brunellexec.NewRunner()
	if err != nil {
		writeResponse(response{ErrorCode: brunellexec.ErrorCode(err), Error: err.Error()})
		os.Exit(1)
	}

	out, runErr := runner.Run(context.Background(), brunellexec.Options{
		Command:        command,
		WorkDir:        workDir,
		Timeout:        timeout,
		MaxProcesses:   uint32(maxProcesses),
		MaxMemoryBytes: maxMemoryBytes,
		MaxOutputBytes: maxOutputBytes,
	})
	result := response{
		Stdout:          string(out.Stdout),
		Stderr:          string(out.Stderr),
		ExitCode:        out.ExitCode,
		StdoutTruncated: out.StdoutTruncated,
		StderrTruncated: out.StderrTruncated,
	}
	if runErr != nil {
		result.ErrorCode = brunellexec.ErrorCode(runErr)
		result.Error = runErr.Error()
	}
	writeResponse(result)
	if runErr != nil {
		os.Exit(1)
	}
}

func writeResponse(result response) {
	if err := json.NewEncoder(os.Stdout).Encode(result); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
	}
}
