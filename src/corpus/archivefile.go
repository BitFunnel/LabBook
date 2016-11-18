package corpus

import (
	"io"
	"io/ioutil"

	"github.com/BitFunnel/LabBook/src/signature"
)

// TODO: Use validation step to populate a `Chunk.path` member?

// ArchiveFile represents a tar'd file that contains a subset of the corpus.
// The SHA512 hash is used to verify the version of the data is correct.
type ArchiveFile struct {
	Name   string `yaml:"name"`
	SHA512 string `yaml:"sha512"`
}

func (chunk *ArchiveFile) validate(reader io.Reader) bool {
	stream, readErr := ioutil.ReadAll(reader)
	if readErr != nil {
		return false
	}

	return signature.ValidateData(stream, chunk.SHA512)
}
