//go:build windows

package config

import "testing"

func TestPlatformCredentialWriterRejectsInvalidValues(t *testing.T) {
	writer := NewPlatformCredentialWriter()
	for _, key := range []string{"", "   ", "with\x00nul", string([]byte{0xff})} {
		if err := writer.SetOpenRouterAPIKey(key); err == nil {
			t.Fatalf("SetOpenRouterAPIKey(%q) unexpectedly succeeded", key)
		}
	}
}
