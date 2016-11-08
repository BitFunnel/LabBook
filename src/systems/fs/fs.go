package fs

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/BitFunnel/LabBook/src/systems"
)

type fsOperation struct {
	opString string
}

func (fsOp fsOperation) String() string {
	return fmt.Sprintf("[FS] %s", fsOp.opString)
}

func newFsOperation(fsOp string) systems.Operation {
	return fsOperation{opString: fsOp}
}

// Chdir is a mockable wrapper for `os.Chdir`.
func Chdir(dir string) error {
	if systems.IsDryRun() {
		// NOTE: This is not a potentially deleterious operaiton, so we don't
		// return early.
		operationText := fmt.Sprintf(`os.Chdir("%s")`, dir)
		operation := newFsOperation(operationText)
		systems.OpLog().Log(&operation)
	}

	return os.Chdir(dir)

}

// MkdirAll is a mockable wrapper for `os.MkdirAll`.
func MkdirAll(path string, perm os.FileMode) error {
	if systems.IsDryRun() {
		operationText := fmt.Sprintf(`os.MkdirAll("%s", 0%o)`, path, perm)
		operation := newFsOperation(operationText)
		systems.OpLog().Log(&operation)

		return nil
	}

	return os.MkdirAll(path, perm)
}

// Create is a mockable wrapper for `os.Create`.
func Create(name string) (*os.File, error) {
	if systems.IsDryRun() {
		operationText := fmt.Sprintf(`os.Create("%s")`, name)
		operation := newFsOperation(operationText)
		systems.OpLog().Log(&operation)

		return os.OpenFile(os.DevNull, os.O_WRONLY, 0777)
	}

	return os.Create(name)
}

// WriteFile is a mockable wrapper for `ioutil.WriteFile`.
func WriteFile(filename string, data []byte, perm os.FileMode) error {
	if systems.IsDryRun() {
		operationText := fmt.Sprintf(`ioutil.WriteFile("%s", ..., 0%o)`, filename, perm)
		operation := newFsOperation(operationText)
		systems.OpLog().Log(&operation)

		return nil
	}

	return ioutil.WriteFile(filename, data, perm)
}
