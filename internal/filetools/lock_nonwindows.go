//go:build !windows

package filetools

import (
	"os"
	"syscall"
)

// openLockable opens path for reading; the handle is used to hold an
// advisory whole-file lock across the stale-hash check and replacement.
func openLockable(path string) (*os.File, error) {
	return os.Open(path)
}

// lockFileExclusive takes a blocking exclusive advisory lock on f.
func lockFileExclusive(f *os.File) error {
	return syscall.Flock(int(f.Fd()), syscall.LOCK_EX)
}

// unlockFile releases the lock taken by lockFileExclusive.
func unlockFile(f *os.File) error {
	return syscall.Flock(int(f.Fd()), syscall.LOCK_UN)
}
