//go:build windows

package exec

import (
	"testing"
	"unsafe"
)

// These sizes are documented, stable Win32 ABI facts (x64) for
// JOBOBJECT_BASIC_LIMIT_INFORMATION (0x40) and
// JOBOBJECT_EXTENDED_LIMIT_INFORMATION (0x90). This repo has no
// golang.org/x/sys/windows dependency to cross-check struct layout
// against, so this test is the regression guard against a field-order
// or type mistake silently producing the wrong memory layout.
func TestJobObjectStructSizes(t *testing.T) {
	if got, want := unsafe.Sizeof(jobObjectBasicLimitInformation{}), uintptr(0x40); got != want {
		t.Fatalf("sizeof(jobObjectBasicLimitInformation) = %#x, want %#x", got, want)
	}
	if got, want := unsafe.Sizeof(jobObjectExtendedLimitInformationT{}), uintptr(0x90); got != want {
		t.Fatalf("sizeof(jobObjectExtendedLimitInformationT) = %#x, want %#x", got, want)
	}
}
