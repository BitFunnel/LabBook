package bfrepo

import (
	"fmt"

	"github.com/BitFunnel/LabBook/src/systems"
)

type bfOperation struct {
	opString string
}

func (bfOp *bfOperation) String() string {
	return fmt.Sprintf("[BitFunnel]    %s", bfOp.opString)
}

func newBfOperation(bfOp string) systems.Operation {
	return &bfOperation{opString: bfOp}
}
