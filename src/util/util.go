package util

import "os"

// Exists checks whether a path exists on the filesystem.
func Exists(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}

	return true
}
