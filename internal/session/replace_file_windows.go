//go:build windows

package session

import (
	"errors"
	"syscall"
	"unsafe"
)

const (
	moveFileReplaceExisting = 0x00000001
	moveFileWriteThrough    = 0x00000008
)

var (
	kernel32          = syscall.NewLazyDLL("kernel32.dll")
	replaceFileW      = kernel32.NewProc("ReplaceFileW")
	moveFileExW       = kernel32.NewProc("MoveFileExW")
	errorFileNotFound = syscall.Errno(2)
	errorPathNotFound = syscall.Errno(3)
)

func replaceExistingFile(source, destination string) error {
	sourceName, err := syscall.UTF16PtrFromString(source)
	if err != nil {
		return err
	}
	destinationName, err := syscall.UTF16PtrFromString(destination)
	if err != nil {
		return err
	}
	result, _, callErr := replaceFileW.Call(
		uintptr(unsafe.Pointer(destinationName)),
		uintptr(unsafe.Pointer(sourceName)),
		0,
		0,
		0,
		0,
	)
	if result != 0 {
		return nil
	}
	if !errors.Is(callErr, errorFileNotFound) && !errors.Is(callErr, errorPathNotFound) {
		return callErr
	}
	result, _, callErr = moveFileExW.Call(
		uintptr(unsafe.Pointer(sourceName)),
		uintptr(unsafe.Pointer(destinationName)),
		moveFileReplaceExisting|moveFileWriteThrough,
	)
	if result == 0 {
		return callErr
	}
	return nil
}
