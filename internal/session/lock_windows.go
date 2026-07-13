//go:build windows

package session

import (
	"errors"
	"syscall"
	"unsafe"
)

const (
	lockfileFailImmediately = 0x00000001
	lockfileExclusiveLock   = 0x00000002
	errorLockViolation      = syscall.Errno(33)
)

var (
	kernel32Lock   = syscall.NewLazyDLL("kernel32.dll")
	lockFileEx     = kernel32Lock.NewProc("LockFileEx")
	unlockFileEx   = kernel32Lock.NewProc("UnlockFileEx")
	errSessionBusy = errors.New("session lock is held")
)

type sessionLock struct {
	handle syscall.Handle
}

func acquireSessionLock(dir string) (sessionLock, error) {
	path, err := syscall.UTF16PtrFromString(dir + `\owner.lock`)
	if err != nil {
		return sessionLock{}, err
	}
	handle, err := syscall.CreateFile(path, syscall.GENERIC_READ|syscall.GENERIC_WRITE, syscall.FILE_SHARE_READ|syscall.FILE_SHARE_WRITE|syscall.FILE_SHARE_DELETE, nil, syscall.OPEN_ALWAYS, syscall.FILE_ATTRIBUTE_NORMAL, 0)
	if err != nil {
		return sessionLock{}, err
	}
	var overlapped syscall.Overlapped
	result, _, callErr := lockFileEx.Call(uintptr(handle), lockfileFailImmediately|lockfileExclusiveLock, 0, 1, 0, uintptr(unsafe.Pointer(&overlapped)))
	if result == 0 {
		_ = syscall.CloseHandle(handle)
		if errors.Is(callErr, errorLockViolation) {
			return sessionLock{}, errSessionBusy
		}
		if callErr == nil {
			callErr = syscall.EINVAL
		}
		return sessionLock{}, callErr
	}
	return sessionLock{handle: handle}, nil
}

func (l sessionLock) Close() error {
	if l.handle == 0 {
		return nil
	}
	var overlapped syscall.Overlapped
	result, _, callErr := unlockFileEx.Call(uintptr(l.handle), 0, 1, 0, uintptr(unsafe.Pointer(&overlapped)))
	closeErr := syscall.CloseHandle(l.handle)
	if result == 0 {
		if callErr == nil {
			callErr = syscall.EINVAL
		}
		return callErr
	}
	return closeErr
}

func sessionLockError(err error) error {
	if errors.Is(err, errSessionBusy) {
		return codeError(ErrSessionBusy.Code, "session is already active", err)
	}
	return codeError("E_SESSION_STORAGE", "cannot acquire session lock", err)
}
