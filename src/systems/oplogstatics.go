package systems

import (
	"fmt"
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

// ConfigureAsTraceRun sets the system to emit tracing information for selected
// OS calls.
func ConfigureAsTraceRun() {
	configureAsTraceRunOnce.Do(func() {
		traceRun = true
	})
}

// IsTraceRun reports whether this is meant to be a tracing run.
func IsTraceRun() bool {
	return traceRun
}

//
// VERBOSE STATICS CONFIGURATION.
//

var verboseRun = false
var configureAsVerboseRunOnce sync.Once
var outputFile *os.File

// ConfigureAsVerboseRun sets the system to pipe all the output of shell'd
// commands to stdout.
func ConfigureAsVerboseRun() {
	configureAsVerboseRunOnce.Do(func() {
		verboseRun = true
		outputFile = os.Stdout
	})
}

// IsVerboseRun reports whether this is meant to be a verbose run.
func IsVerboseRun() bool {
	return verboseRun
}

// OutputFile returns the file to write the output of shell commands to.
func OutputFile() *os.File {
	return outputFile
}

//
// STATICS INITIALIZATION.
//

func init() {
	nullOutputFile, openErr := os.OpenFile(os.DevNull, os.O_WRONLY, 0777)
	if openErr != nil {
		panicMessage := fmt.Sprintf("Failed to configure LabBook: could not open '%s'", os.DevNull)
		panic(panicMessage)
	}
	outputFile = nullOutputFile

	opLog = &operationLog{
		logger:   log.New(os.Stdout, "", 0),
		eventLog: []Operation{},
	}
}
