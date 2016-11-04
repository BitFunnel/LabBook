package errors

import (
	"log"
	"os"
)

var Logger = log.New(os.Stderr, "", 0)

// CheckFatal will log `err` and an error message `msg` if an error is present,
// and then exit the process.
func CheckFatal(err error, msg string) {
	if err != nil {
		Logger.Fatalf("%s:\n%v", msg, err)
	}
}

// CheckFatalB will log `err` if an error is present, and then exit the process.
func CheckFatalB(err error) {
	if err != nil {
		Logger.Fatalf("%v", err)
	}
}
