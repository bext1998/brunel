//go:build !windows

package filetools

import "os"

func replaceExistingFile(source, destination string) error {
	return os.Rename(source, destination)
}
