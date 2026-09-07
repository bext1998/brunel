//go:build windows

package filetools

import (
	"errors"
	"os"
	"syscall"
	"unsafe"
)

var (
	createFileW = kernel32.NewProc("CreateFileW")
	lockFileEx  = kernel32.NewProc("LockFileEx")
	unlockFileX = kernel32.NewProc("UnlockFileEx")
)

const (
	genericRead             = 0x80000000
	fileShareRead           = 0x00000001
	fileShareWrite          = 0x00000002
	fileShareDelete         = 0x00000004
	openExisting            = 3
	fileAttributeNorm       = 0x80
	lockfileFailImmediately = 0x00000001
	lockfileExclusive       = 0x00000002
	errorLockViolation      = syscall.Errno(33)
	// lockWholeFileLow/High cover the entire file: LockFileEx over a
	// maximal byte range is the conventional whole-file exclusive lock.
	lockWholeFileLow  = 0xFFFFFFFF
	lockWholeFileHigh = 0x7FFFFFFF
)

// openLockable opens path for reading with full sharing - including
// FILE_SHARE_DELETE - so the returned handle can hold an exclusive
// byte-range lock while replaceExistingFile renames the file out from
// under it.
func openLockable(path string) (*os.File, error) {
	pathPtr, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return nil, err
	}
	handle, _, callErr := createFileW.Call(
		uintptr(unsafe.Pointer(pathPtr)),
		uintptr(genericRead),
		uintptr(fileShareRead|fileShareWrite|fileShareDelete),
		0,
		uintptr(openExisting),
		uintptr(fileAttributeNorm),
		0,
	)
	if handle == uintptr(syscall.InvalidHandle) {
		return nil, callErr
	}
	return os.NewFile(handle, path), nil
}

// lockFileExclusive attempts a non-blocking exclusive whole-file lock on f.
func lockFileExclusive(f *os.File) error {
	overlapped := syscall.Overlapped{}
	result, _, callErr := lockFileEx.Call(
		f.Fd(),
		uintptr(lockfileExclusive|lockfileFailImmediately),
		0,
		uintptr(lockWholeFileLow),
		uintptr(lockWholeFileHigh),
		uintptr(unsafe.Pointer(&overlapped)),
	)
	if result == 0 {
		return callErr
	}
	return nil
}

func isLockUnavailable(err error) bool {
	return errors.Is(err, errorLockViolation)
}

// unlockFile releases the whole-file lock taken by lockFileExclusive.
func unlockFile(f *os.File) error {
	overlapped := syscall.Overlapped{}
	result, _, callErr := unlockFileX.Call(
		f.Fd(),
		0,
		uintptr(lockWholeFileLow),
		uintptr(lockWholeFileHigh),
		uintptr(unsafe.Pointer(&overlapped)),
	)
	if result == 0 {
		return callErr
	}
	return nil
}
