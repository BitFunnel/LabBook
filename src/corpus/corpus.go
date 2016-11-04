package corpus

import (
	"crypto/sha512"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"

	"github.com/bitfunnel/LabBook/src/cmd"
	"github.com/bitfunnel/LabBook/src/util"
)

// Manager is responsible for the lifecycle of the corpus, including
// downloading, verifying, and uncompressing.
type Manager interface {
	Uncompress() error
}

// Chunk represents a tar'd file that contains a subset of the corpus. the
// SHA512 hash is used to verify the version of the data is correct.
type Chunk struct {
	Name   string `yaml:"name"`
	SHA512 string `yaml:"sha512"`
}

// New makes a `Manager`, which can be used to govern the lifecycle of a
// corpus directory.
func New(chunks []Chunk, corpusRoot string) Manager {
	return corpusContext{chunks: chunks, corpusRoot: corpusRoot}
}

func (ctx corpusContext) Uncompress() error {
	for _, chunk := range ctx.chunks {
		if !strings.HasSuffix(chunk.Name, ".tar.gz") {
			return fmt.Errorf("Corpus file '%s' is not a .tar.gz file",
				chunk.Name)
		}

		chunkPath := fmt.Sprintf("%s/%s", ctx.corpusRoot, chunk.Name)

		if !util.Exists(chunkPath) {
			return fmt.Errorf("Corpus file '%s' does not exist.", chunkPath)
		}

		chunkFile, openErr := os.Open(chunkPath)
		defer chunkFile.Close()
		if openErr != nil {
			return openErr
		}

		if !chunk.validate(chunkFile) {
			return fmt.Errorf("SHA512 hash for corpus file '%s' does not "+
				"match the hash specified in experiment YAML", chunkPath)
		}

		// TODO: Probably we can avoid un-taring this all the time, and also
		// use a pure solution.
		tarErr := cmd.RunCommand("tar", "-xf", chunkPath, "-C", ctx.corpusRoot)
		if tarErr != nil {
			return tarErr
		}
	}

	return nil
}

func (chunk Chunk) validate(reader io.Reader) bool {
	stream, readErr := ioutil.ReadAll(reader)
	if readErr != nil {
		return false
	}

	hash := sha512.New()
	hash.Write(stream)
	actualSha512Hash := fmt.Sprintf("%x", hash.Sum(nil))

	return strings.ToLower(actualSha512Hash) ==
		strings.ToLower(chunk.SHA512)
}

type corpusContext struct {
	chunks     []Chunk
	corpusRoot string
}
