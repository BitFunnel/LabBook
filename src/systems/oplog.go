package systems

import "log"

// Operation is a single constituent of the `OperationLog`. Typically a systems
// operation (like a filesystem write) is emitted as an `Operation` to the
// `OperationLog` so that systems operations can be audited and traced.
type Operation interface {
	String() string
}

// OperationLog is a log of operations. Used to do things like trace filesystem
// calls.
type OperationLog interface {
	Log(op Operation)
	GetEventLog() []Operation
	ResetEventLog()
}

type operationLog struct {
	logger   *log.Logger
	eventLog []Operation
}

func (log *operationLog) Log(op Operation) {
	log.eventLog = append(log.eventLog, op)
	if !testRun {
		log.logger.Printf(op.String())
	}
}

func (log *operationLog) GetEventLog() []Operation {
	return log.eventLog
}

func (log *operationLog) ResetEventLog() {
	log.eventLog = []Operation{}
}
