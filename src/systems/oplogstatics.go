package systems

import (
	"log"
	"os"
	"sync"
)

var opLog OperationLog

// OpLog is the canonical systems `OperationLog`, used to log operations from
// subsystems like the filesystem.
func OpLog() OperationLog {
	return opLog
}

//
// DRY RUN STATICS CONFIGURATION.
//

var dryRun = false
var configureAsDryRunOnce sync.Once

// ConfigureAsDryRun sets the system to not perform deleterious effects.
func ConfigureAsDryRun() {
	configureAsDryRunOnce.Do(func() {
		dryRun = true
	})
}

// IsDryRun reports whether this is meant to be a dry run (i.e., whether
// potentially deleterious operations will be logged but not executed).
func IsDryRun() bool {
	return dryRun
}

//
// TEST RUN STATICS CONFIGURATION.
//

var testRun = false
var configureAsTestRunOnce sync.Once

// ConfigureAsTestRun sets the system to not perform deleterious effects.
func ConfigureAsTestRun() {
	configureAsTestRunOnce.Do(func() {
		testRun = true
	})
	ConfigureAsDryRun()
}

// IsTestRun reports whether this is meant to be a test run.
func IsTestRun() bool {
	return testRun
}

//
// STATICS INITIALIZATION.
//

func init() {
	opLog = &operationLog{
		logger:   log.New(os.Stdout, "", 0),
		eventLog: []Operation{},
	}
}
