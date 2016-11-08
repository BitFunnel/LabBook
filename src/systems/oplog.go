package systems

import (
	"log"
	"os"
	"sync"
)

// Operation is a single constituent of the `OperationLog`. Typically a systems
// operation (like a filesystem write) is emitted as an `Operation` to the
// `OperationLog` so that systems operations can be audited and traced.
type Operation interface {
	String() string
}

// OperationLog is a log of operations. Used to do things like trace filesystem
// calls.
type OperationLog interface {
	Log(op *Operation)
}

type operationLog struct {
	logger *log.Logger
}

func (log *operationLog) Log(op *Operation) {
	log.logger.Printf((*op).String())
}

var opLog OperationLog

// OpLog is the canonical systems `OperationLog`, used to log operations from
// subsystems like the filesystem.
func OpLog() OperationLog {
	return opLog
}

var dryRun = false
var configureOnce sync.Once

// ConfigureAsDryRun sets the system to not perform deleterious effects.
func ConfigureAsDryRun() {
	configureOnce.Do(func() {
		dryRun = true
	})
}

// IsDryRun reports whether this is meant to be a dry run (i.e., whether
// potentially deleterious operations will be logged but not executed).
func IsDryRun() bool {
	return dryRun
}

func init() {
	opLog = &operationLog{logger: log.New(os.Stdout, "", 0)}
}
