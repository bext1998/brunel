//go:build !windows

package session

import (
	"errors"
	"os"
	"syscall"
)

var errSessionBusy = errors.New("session lock is held")

type sessionLock struct {
	file *os.File
}

func acquireSessionLock(dir string) (sessionLock, error) {
	file, err := os.OpenFile(dir+string(os.PathSeparator)+"owner.lock", os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		return sessionLock{}, err
	}
	if err := syscall.Flock(int(file.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		_ = file.Close()
		if errors.Is(err, syscall.EWOULDBLOCK) || errors.Is(err, syscall.EAGAIN) {
			return sessionLock{}, errSessionBusy
		}
		return sessionLock{}, err
	}
	return sessionLock{file: file}, nil
}

func (l sessionLock) Close() error {
	if l.file == nil {
		return nil
	}
	return l.file.Close()
}

func sessionLockError(err error) error {
	if errors.Is(err, errSessionBusy) {
		return codeError(ErrSessionBusy.Code, "session is already active", err)
	}
	return codeError("E_SESSION_STORAGE", "cannot acquire session lock", err)
}
