package traceablefs

import (
	"fmt"

	"github.com/BitFunnel/LabBook/src/systems"
)

type fsOperation struct {
	opString string
}

func (fsOp *fsOperation) String() string {
	return fmt.Sprintf("[TRACEABLE FS]\t%s", fsOp.opString)
}

func newFsOperation(fsOp string) systems.Operation {
	return &fsOperation{opString: fsOp}
}
