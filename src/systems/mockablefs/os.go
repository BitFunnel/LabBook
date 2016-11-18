package mockablefs

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/BitFunnel/LabBook/src/systems"
)

// Open is a mockable wrapper for `os.Open`.
func Open(name string) (*os.File, error) {
	if systems.IsDryRun() {
		operationText := fmt.Sprintf(`os.Open("%s")`, name)
		systems.OpLog().Log(newFsOperation(operationText))

		return os.Open(os.DevNull)
	}

	return os.Open(name)
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
