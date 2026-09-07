//go:build !windows

package filetools

import "os"

// installNewFile atomically installs source at destination with
// no-overwrite semantics: os.Rename on POSIX fails when destination
// exists, leaving destination untouched.
func installNewFile(source, destination string) error {
	return os.Rename(source, destination)
}

// isAlreadyExists reports whether err is the "destination exists" signal
// from a no-overwrite rename.
func isAlreadyExists(err error) bool {
	return os.IsExist(err)
}
