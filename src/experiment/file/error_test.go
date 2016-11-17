package file

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_SimpleFileErrors(t *testing.T) {
	destExistsErr := destinationExistsError("/test/path")
	assert.Error(t, destExistsErr)
	assert.True(t, isDestinationExistsError(destExistsErr))
	assert.False(t, isSourceDoesNotExistError(destExistsErr))
	assert.False(t, isCouldNotRemoveSourceError(destExistsErr))

	sourceNotExistsErr := sourceDoesNotExistError("/test/path")
	assert.Error(t, sourceNotExistsErr)
	assert.False(t, isDestinationExistsError(sourceNotExistsErr))
	assert.True(t, isSourceDoesNotExistError(sourceNotExistsErr))
	assert.False(t, isCouldNotRemoveSourceError(sourceNotExistsErr))

	notRemoveSourceErr := couldNotRemoveSourceError(
		"/test/path",
		errors.New("fake error from os.Remove"))
	assert.Error(t, notRemoveSourceErr)
	assert.False(t, isDestinationExistsError(notRemoveSourceErr))
	assert.False(t, isSourceDoesNotExistError(notRemoveSourceErr))
	assert.True(t, isCouldNotRemoveSourceError(notRemoveSourceErr))
}
