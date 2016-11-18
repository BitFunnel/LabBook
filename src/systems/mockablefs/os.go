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

// OpenDo will open a file, retrieve all data from it, perform `action` on that
// data, and then close the file.
func OpenDo(name string, action func(data []byte) error) (err error) {
	file, openErr := Open(name)
	if openErr != nil {
		return openErr
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil && err == nil {
			err = closeErr
		}
	}()

	data, readErr := ioutil.ReadAll(file)
	if readErr != nil {
		return readErr
	}

	return action(data)
}

// OpenDoFile will open a file, perform `action` on that file, and then close
// it.
func OpenDoFile(name string, action func(file *os.File) error) (err error) {
	file, openErr := Open(name)
	if openErr != nil {
		return openErr
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil && err == nil {
			err = closeErr
		}
	}()

	return action(file)
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

// CreateDo will perform `os.Create`, perform `action` on the resulting
// file, and then sync and close that file.
func CreateDo(name string, action func(*os.File) error) (err error) {
	file, createErr := Create(name)
	if createErr != nil {
		return createErr
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil && err == nil {
			err = closeErr
		}
	}()

	if actionErr := action(file); actionErr != nil {
		return actionErr
	}

	return file.Sync()
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
