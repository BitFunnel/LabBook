package util

import "os"

// Exists checks whether a path exists on the filemanager.
func Exists(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}

	return true
}

// IsDir checks if path points at a directory.
func IsDir(path string) bool {
	fileInfo, statErr := os.Stat(path)
	if statErr != nil || !fileInfo.IsDir() {
		return false
	}
	return true
}
