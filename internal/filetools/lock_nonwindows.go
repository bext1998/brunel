//go:build !windows

package filetools

import (
	"errors"
	"os"
	"syscall"
)

// openLockable opens path for reading; the handle is used to hold an
// advisory whole-file lock across the stale-hash check and replacement.
func openLockable(path string) (*os.File, error) {
	return os.Open(path)
}

// lockFileExclusive attempts a non-blocking exclusive advisory lock on f.
func lockFileExclusive(f *os.File) error {
	return syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
}

func isLockUnavailable(err error) bool {
	return errors.Is(err, syscall.EWOULDBLOCK) || errors.Is(err, syscall.EAGAIN)
}

// unlockFile releases the lock taken by lockFileExclusive.
func unlockFile(f *os.File) error {
	return syscall.Flock(int(f.Fd()), syscall.LOCK_UN)
}
