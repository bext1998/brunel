//go:build !windows

package exec

import "context"

// Runner is unavailable outside Brunel's supported Windows platform.
type Runner struct{}

func NewRunner() (*Runner, error) {
	return nil, codeError(ErrUnsupportedPlatform.Code, "PowerShell Job Object execution requires Windows", nil)
}

func (r *Runner) Run(ctx context.Context, opts Options) (Output, error) {
	return Output{}, codeError(ErrUnsupportedPlatform.Code, "PowerShell Job Object execution requires Windows", nil)
}
