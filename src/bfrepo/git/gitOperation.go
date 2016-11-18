package git

import (
	"fmt"

	"github.com/BitFunnel/LabBook/src/systems"
)

type gitOperation struct {
	opString string
}

func (gitOp *gitOperation) String() string {
	return fmt.Sprintf("[GIT]          %s", gitOp.opString)
}

func newGitOperation(gitOp string) systems.Operation {
	return &gitOperation{opString: gitOp}
}
