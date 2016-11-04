package util

import (
	"log"
	"os"
)

// Check will log `err` and an error message `msg` if an error is present, and
// then exit the process.
func Check(err error, msg string) {
	if err != nil {
		log.Fatalf("%s:\n%v", msg, err)
	}
}

// CheckErr will log `err` if an error is present, and then exit the process.
func CheckErr(err error) {
	if err != nil {
		log.Fatalf("%v", err)
	}
}

// Exists checks whether a path exists on the filesystem.
func Exists(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}

	return true
}
