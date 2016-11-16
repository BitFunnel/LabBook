package util

import (
	"crypto/sha512"
	"fmt"
	"os"
	"strings"
)

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

// NormalizeSignature puts a signature string into canonical form.
func NormalizeSignature(signature string) string {
	return strings.ToLower(signature)
}

// TODO: Move ValidateSHA512 out of util and into file/lock.

// ValidateSHA512 will hash a stream of bytes using a canonical SHA512
// configuration, and validate that it matches a given SHA512 hash.
func ValidateSHA512(stream []byte, SHA512 string) bool {
	hash := sha512.New()
	hash.Write(stream)
	actualSha512Hash := fmt.Sprintf("%x", hash.Sum(nil))

	return strings.ToLower(actualSha512Hash) ==
		strings.ToLower(SHA512)
}
