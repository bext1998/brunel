//go:build windows

package filetools

import (
	"errors"
	"syscall"
	"unsafe"
)

const (
	moveFileWriteThroughOnly = 0x00000008 // MOVEFILE_WRITE_THROUGH without REPLACE_EXISTING
)

// installNewFile atomically installs source at destination with
// no-overwrite semantics: it fails when destination already exists,
// leaving destination untouched. MOVEFILE_REPLACE_EXISTING is not passed,
// so MoveFileExW fails with ERROR_ALREADY_EXISTS if destination is
// present.
func installNewFile(source, destination string) error {
	sourceName, err := syscall.UTF16PtrFromString(source)
	if err != nil {
		return err
	}
	destinationName, err := syscall.UTF16PtrFromString(destination)
	if err != nil {
		return err
	}
	result, _, callErr := moveFileExW.Call(
		uintptr(unsafe.Pointer(sourceName)),
		uintptr(unsafe.Pointer(destinationName)),
		uintptr(moveFileWriteThroughOnly),
	)
	if result != 0 {
		return nil
	}
	return callErr
}

// isAlreadyExists reports whether err is Windows ERROR_ALREADY_EXISTS
// (183) or ERROR_FILE_EXISTS (80), the "destination exists" signals from
// a no-overwrite MoveFileExW.
func isAlreadyExists(err error) bool {
	return errors.Is(err, syscall.Errno(183)) || errors.Is(err, syscall.Errno(80))
}
