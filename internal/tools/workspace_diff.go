package tools

import (
	"context"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/bext1998/brunel/internal/filetools"
)

const workspaceDiffTimeout = 30 * time.Second

func workspaceDiff(ctx context.Context, r filetools.Resolver, root, path string) (string, error) {
	// Resolve before spawning git so a path escape is rejected without any I/O.
	if _, err := r.Resolve(path); err != nil {
		return "", err
	}
	pathspec := "."
	if path != "" {
		pathspec = filepath.ToSlash(filepath.Clean(path))
	}
	commandCtx, cancel := context.WithTimeout(ctx, workspaceDiffTimeout)
	defer cancel()
	cmd := exec.CommandContext(commandCtx, "git", "-C", root, "diff", "--", pathspec)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", codeError(ErrWorkspaceDiffUnavailable.Code, "git diff is unavailable for this workspace", err)
	}
	return string(output), nil
}
