//go:build !windows

package config

import (
	"context"
	"errors"
	"testing"
)

func TestPlatformCredentialSourceUnsupported(t *testing.T) {
	loader := NewLoader(t.TempDir(), t.TempDir(), nil)
	_, err := loader.Load(context.Background(), CLIOverrides{})
	if !errors.Is(err, ErrUnsupportedPlatform) {
		t.Fatalf("Load() error = %v, want E_UNSUPPORTED_PLATFORM", err)
	}
}
