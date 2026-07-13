//go:build !windows

package session

import "os"

func replaceExistingFile(source, destination string) error {
	return os.Rename(source, destination)
}
