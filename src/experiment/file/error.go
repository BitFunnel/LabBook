package file

import "fmt"

type lockErrorType uint

const (
	destinationExists lockErrorType = iota
	sourceDoesNotExist
	couldNotRemoveSource
	unknown
)

type lockError struct {
	errorType lockErrorType
	path      string
	innerErr  error
}

func (err *lockError) Error() string {
	switch err.errorType {
	case destinationExists:
		return fmt.Sprintf("Could not acquire lock file at '%s' because a .LOCKFILE in the same directory exists. If you are sure no other process is modifying this directory, move .LOCKFILE -> LOCKFILE and try again. If you think the .LOCKFILE is corrupt, delete and re-generate this directory", err.path)
	case sourceDoesNotExist:
		return fmt.Sprintf("Could not acquire lock file at '%s' because it does not exist", err.path)
	case couldNotRemoveSource:
		return fmt.Sprintf("Could not acquire lock file at '%s'; attempted to move LOCKFILE -> .LOCKFILE atomically and then delete LOCKFILE, but we failed to delete LOCKFILE. If another process moved .LOCKFILE and you are sure it's not corrupt, try moving .LOCKFILE -> LOCKFILE and try again. The delete error follows:\n%v", err.path, err.innerErr)
	case unknown:
		return fmt.Sprintf("Could not acquire lock file at '%s'; attempted to move LOCKFILE -> .LOCKFILE, but failed with error:\n%v", err.path, err.innerErr)
	default:
		return fmt.Sprintf("Could not acquire lock file, but error type is unknown (this should not happen, please file a bug)")
	}
}

func isDestinationExistsError(err error) bool {
	switch pe := err.(type) {
	case nil:
		return false
	case *lockError:
		return pe.errorType == destinationExists
	default:
		return false
	}
}

func isSourceDoesNotExistError(err error) bool {
	switch pe := err.(type) {
	case nil:
		return false
	case *lockError:
		return pe.errorType == sourceDoesNotExist
	default:
		return false
	}
}

func isCouldNotRemoveSourceError(err error) bool {
	switch pe := err.(type) {
	case nil:
		return false
	case *lockError:
		return pe.errorType == couldNotRemoveSource
	default:
		return false
	}
}

func destinationExistsError(path string) error {
	return &lockError{errorType: destinationExists, path: path}
}

func sourceDoesNotExistError(path string) error {
	return &lockError{errorType: sourceDoesNotExist, path: path}
}

func couldNotRemoveSourceError(path string, innerErr error) error {
	return &lockError{
		errorType: couldNotRemoveSource,
		path:      path,
		innerErr:  innerErr,
	}
}

func unknownLockError(path string, innerErr error) error {
	return &lockError{
		errorType: unknown,
		path:      path,
		innerErr:  innerErr,
	}
}
