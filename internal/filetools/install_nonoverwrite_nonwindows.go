//go:build !windows

package filetools

import "os"

// installNewFile atomically installs source at destination with
// no-overwrite semantics. Linking fails with EEXIST when destination
// already exists; removing the temporary source afterwards leaves the new
// destination in place. source and destination always share a directory.
func installNewFile(source, destination string) error {
	if err := os.Link(source, destination); err != nil {
		return err
	}
	_ = os.Remove(source)
	return nil
}

// isAlreadyExists reports whether err is the EEXIST signal from a
// no-overwrite link.
func isAlreadyExists(err error) bool {
	return os.IsExist(err)
}
