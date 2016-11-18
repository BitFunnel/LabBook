package traceablefs

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/BitFunnel/LabBook/src/systems"
	"github.com/BitFunnel/LabBook/src/systems/shell"
)

// Open is a traceable wrapper for `os.Open`.
func Open(name string) (*os.File, error) {
	if systems.IsTraceRun() {
		operationText := fmt.Sprintf(`os.Open("%s")`, name)
		systems.OpLog().Log(newFsOperation(operationText))
	}

	return os.Open(name)
}

// Remove is a traceable wrapper for `os.Remove`.
func Remove(name string) error {
	if systems.IsTraceRun() {
		operationText := fmt.Sprintf(`os.Remove("%s")`, name)
		systems.OpLog().Log(newFsOperation(operationText))
	}

	return os.Remove(name)
}

// Link is a traceable wrapper for `os.Link`.
func Link(oldname, newname string) error {
	if systems.IsTraceRun() {
		operationText := fmt.Sprintf(`os.Link("%s", "%s")`, oldname, newname)
		systems.OpLog().Log(newFsOperation(operationText))
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

// MkdirAll is a traceable wrapper for `os.MkdirAll`.
func MkdirAll(path string, perm os.FileMode) error {
	if systems.IsTraceRun() {
		operationText := fmt.Sprintf(`os.MkdirAll("%s", 0%o)`, path, perm)
		systems.OpLog().Log(newFsOperation(operationText))
	}

	return os.MkdirAll(path, perm)
}

// Create is a traceable wrapper for `os.Create`.
func Create(name string) (*os.File, error) {
	if systems.IsTraceRun() {
		operationText := fmt.Sprintf(`os.Create("%s")`, name)
		systems.OpLog().Log(newFsOperation(operationText))
	}

	return os.Create(name)
}

// WriteFile is a traceable wrapper for `ioutil.WriteFile`.
func WriteFile(filename string, data []byte, perm os.FileMode) error {
	if systems.IsTraceRun() {
		operationText := fmt.Sprintf(`ioutil.WriteFile("%s", ..., 0%o)`, filename, perm)
		systems.OpLog().Log(newFsOperation(operationText))
	}

	return ioutil.WriteFile(filename, data, perm)
}

// chdir is a traceable wrapper for `os.Chdir`.
func chdir(dir string) error {
	if systems.IsTraceRun() {
		operationText := fmt.Sprintf(`os.Chdir("%s")`, dir)
		systems.OpLog().Log(newFsOperation(operationText))
	}

	return os.Chdir(dir)
}
