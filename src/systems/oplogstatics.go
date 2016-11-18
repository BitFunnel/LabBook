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
	ConfigureAsTraceRun()
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
// TRACEABLE STATICS CONFIGURATION.
//

var traceRun = false
var configureAsTraceRunOnce sync.Once

// ConfigureAsTraceRun sets the system to not perform deleterious effects.
func ConfigureAsTraceRun() {
	configureAsTraceRunOnce.Do(func() {
		traceRun = true
	})
}

// IsTraceRun reports whether this is meant to be a test run.
func IsTraceRun() bool {
	return traceRun
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
