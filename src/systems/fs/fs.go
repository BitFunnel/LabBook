package fs

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/BitFunnel/LabBook/src/systems"
	"github.com/BitFunnel/LabBook/src/systems/shell"
)

type fsOperation struct {
	opString string
}

func (fsOp *fsOperation) String() string {
	return fmt.Sprintf("[FS] %s", fsOp.opString)
}

func newFsOperation(fsOp string) systems.Operation {
	return &fsOperation{opString: fsOp}
}

// Open is a mockable wrapper for `os.Open`.
func Open(name string) (*os.File, error) {
	if systems.IsDryRun() {
		operationText := fmt.Sprintf(`os.Open("%s")`, name)
		systems.OpLog().Log(newFsOperation(operationText))

		return os.Open(os.DevNull)
	}

	return os.Open(name)
}

// Remove is a mockable wrapper for `os.Remove`.
func Remove(name string) error {
	if systems.IsDryRun() {
		operationText := fmt.Sprintf(`os.Remove("%s")`, name)
		systems.OpLog().Log(newFsOperation(operationText))

		// TODO: Figure out the semantics of the dry run here.
	}

	return os.Remove(name)
}

// Link is a mockable wrapper for `os.Link`.
func Link(oldname, newname string) error {
	if systems.IsDryRun() {
		operationText := fmt.Sprintf(`os.Link("%s", "%s")`, oldname, newname)
		systems.OpLog().Log(newFsOperation(operationText))

		// TODO: Figure out the semantics of the dry run here.
	}

	return os.Link(oldname, newname)
}

// ScopedChdir changes to `directory` and then, when `Dispose` is
// called, it changes back to the current working directory.
func ScopedChdir(directory string) (shell.CmdHandle, error) {
	pwd, pwdErr := os.Getwd()
	if pwdErr != nil {
		return nil, pwdErr
	}

	chdirErr := chdir(directory)
	if chdirErr != nil {
		return nil, chdirErr
	}

	return shell.MakeHandle(func() error { return chdir(pwd) }), nil
}

// MkdirAll is a mockable wrapper for `os.MkdirAll`.
func MkdirAll(path string, perm os.FileMode) error {
	if systems.IsDryRun() {
		operationText := fmt.Sprintf(`os.MkdirAll("%s", 0%o)`, path, perm)
		systems.OpLog().Log(newFsOperation(operationText))

		return nil
	}

	return os.MkdirAll(path, perm)
}

// Create is a mockable wrapper for `os.Create`.
func Create(name string) (*os.File, error) {
	if systems.IsDryRun() {
		operationText := fmt.Sprintf(`os.Create("%s")`, name)
		systems.OpLog().Log(newFsOperation(operationText))

		return os.OpenFile(os.DevNull, os.O_WRONLY, 0777)
	}

	return os.Create(name)
}

// WriteFile is a mockable wrapper for `ioutil.WriteFile`.
func WriteFile(filename string, data []byte, perm os.FileMode) error {
	if systems.IsDryRun() {
		operationText := fmt.Sprintf(`ioutil.WriteFile("%s", ..., 0%o)`, filename, perm)
		systems.OpLog().Log(newFsOperation(operationText))

		return nil
	}

	return ioutil.WriteFile(filename, data, perm)
}

// chdir is a mockable wrapper for `os.Chdir`.
func chdir(dir string) error {
	if systems.IsDryRun() {
		// NOTE: This is not a potentially deleterious operaiton, so we don't
		// return early.
		operationText := fmt.Sprintf(`os.Chdir("%s")`, dir)
		systems.OpLog().Log(newFsOperation(operationText))
	}

	return os.Chdir(dir)
}
