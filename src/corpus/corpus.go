package corpus

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/BitFunnel/LabBook/src/systems/shell"
	"github.com/BitFunnel/LabBook/src/util"
)

// Manager is responsible for the lifecycle of the corpus, including
// downloading, verifying, and uncompressing.
type Manager interface {
	Uncompress() error
	GetAllCorpusFilepaths() ([]string, error)
}

// NewManager makes a `Manager`, which can be used to govern the lifecycle of a
// corpus directory.
func NewManager(chunks []*Chunk, corpusRoot string) Manager {
	return &corpusContext{
		chunks:       chunks,
		corpusRoot:   corpusRoot,
		uncompressed: false,
	}
}

// Uncompress will uncompress the corpus files that `ctx` is responsible for
// managing.
func (ctx *corpusContext) Uncompress() error {
	if ctx.uncompressed {
		return fmt.Errorf("Corpus at '%s' has already been uncompressed",
			ctx.corpusRoot)
	}

	for _, chunk := range ctx.chunks {
		if !strings.HasSuffix(chunk.Name, ".tar.gz") {
			return fmt.Errorf("Corpus file '%s' is not a .tar.gz file",
				chunk.Name)
		}

		chunkPath := filepath.Join(ctx.corpusRoot, chunk.Name)

		if !util.Exists(chunkPath) {
			return fmt.Errorf("Corpus file '%s' does not exist.", chunkPath)
		}

		chunkFile, openErr := os.Open(chunkPath)
		if openErr != nil {
			return openErr
		}
		defer chunkFile.Close()

		if !chunk.validate(chunkFile) {
			return fmt.Errorf("SHA512 hash for corpus file '%s' does not "+
				"match the hash specified in experiment YAML", chunkPath)
		}

		// TODO: Probably we can avoid un-taring this all the time, and also
		// use a pure solution.
		tarErr := shell.RunCommand("tar", "-xf", chunkPath, "-C", ctx.corpusRoot)
		if tarErr != nil {
			return tarErr
		}
	}

	ctx.uncompressed = true

	return nil
}

// GetAllCorpusFilepaths returns the absolute path of every file in the corpus.
func (ctx *corpusContext) GetAllCorpusFilepaths() ([]string, error) {
	if !ctx.uncompressed {
		return []string{}, fmt.Errorf("Can't get paths of corpus files "+
			"rooted at '%s', since they haven't been uncompressed yet",
			ctx.corpusRoot)
	}

	corpusFiles := []string{}
	files, lsErr := ioutil.ReadDir(ctx.corpusRoot)
	if lsErr != nil {
		return []string{}, fmt.Errorf("Attempted to scan corpus directory '%s' for corpus files, but failed:\n%v", ctx.corpusRoot, lsErr)
	}

	// Obtain paths to all the corpus files. We expect every the corpus root to
	// contain only tarballs (which are the corpus) and folders (which were
	// generated when we called `Uncompress`). IMPORTANT: Any file we find in
	// these subfolders is considered a corpus file.
	for _, file := range files {
		// Corpus root should contains tarballs (i.e, compressed corpus files)
		// or directories. Skip the tarballs.
		if !file.IsDir() {
			continue
		}

		// Recursively look for all non-directory folders in the corpus root.
		// We consider each file to be a corpus file; if it's a file, but it's
		// not a part of the corpus, it doesn't belong in the corpus
		// directories!
		absoluteDirectoryPath := filepath.Join(ctx.corpusRoot, file.Name())
		walkErr := filepath.Walk(
			absoluteDirectoryPath,
			func(path string, fileInfo os.FileInfo, err error) error {
				return corpusFileVisitor(
					&corpusFiles,
					path,
					fileInfo,
					err)
			})
		if walkErr != nil {
			return []string{}, walkErr
		}
	}

	return corpusFiles, nil
}

func corpusFileVisitor(corpusFiles *[]string, path string, fileInfo os.FileInfo, err error) error {
	if err != nil {
		return err
	}

	if !fileInfo.IsDir() {
		*corpusFiles = append(*corpusFiles, path)
	}

	return nil
}

func (chunk Chunk) validate(reader io.Reader) bool {
	stream, readErr := ioutil.ReadAll(reader)
	if readErr != nil {
		return false
	}

	return util.ValidateSHA512(stream, chunk.SHA512)
}

type corpusContext struct {
	chunks       []*Chunk
	corpusRoot   string
	uncompressed bool
}

// Chunk represents a tar'd file that contains a subset of the corpus. the
// SHA512 hash is used to verify the version of the data is correct.
type Chunk struct {
	Name   string `yaml:"name"`
	SHA512 string `yaml:"sha512"`
}
