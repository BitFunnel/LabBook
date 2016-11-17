package corpus

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/BitFunnel/LabBook/src/signature"
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

// TODO: Rename `Uncompress` -> `Decompress`.

// Uncompress will uncompress the corpus files that `ctx` is responsible for
// managing.
func (ctx *corpusContext) Uncompress() error {
	if ctx.uncompressed {
		return fmt.Errorf("Corpus at '%s' has already been uncompressed",
			ctx.corpusRoot)
	}

	// TODO: Implement locking protocol:
	// * Acquire lock. When you're finished uncompressing, write new lock with
	//   the signature.

	signatureAccumulator := signature.NewCorpusSignatureAccumulator()

	for _, chunk := range ctx.chunks {
		chunkPath := ctx.getChunkPath(chunk)

		tarballData, readErr := getCompressedChunkData(chunkPath)
		if readErr != nil {
			return readErr
		}

		tarballSignature, sigErr :=
			signatureAccumulator.AddCorpusTarball(tarballData)
		if sigErr != nil {
			return sigErr
		} else if tarballSignature != chunk.SHA512 {
			return fmt.Errorf("Signature for corpus file '%s' does not "+
				"match the hash specified in experiment YAML; it is "+
				"possible you have specified an incorrect corpus file",
				chunkPath)
		}

		// TODO: Probably we can avoid un-taring this all the time, and also
		// use a pure solution.
		tarErr := shell.RunCommand(
			"tar",
			"-xf",
			chunkPath,
			"-C",
			ctx.corpusRoot)
		if tarErr != nil {
			return tarErr
		}
	}

	// TODO: Have this return the signature.
	_, sigErr := signatureAccumulator.Signature()
	if sigErr != nil {
		return sigErr
	}

	ctx.uncompressed = true

	return nil
}

// GetAllCorpusFilepaths returns the absolute path of every file in the corpus.
func (ctx *corpusContext) GetAllCorpusFilepaths() ([]string, error) {
	// TODO: Consider making chunk files have a `.chunk` suffix to simplify
	// this.
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

//
// PRIVATE METHODS.
//
func (ctx *corpusContext) getChunkPath(chunk *Chunk) string {
	return filepath.Join(ctx.corpusRoot, chunk.Name)
}

//
// PRIVATE FUNCTIONS.
//

func getCompressedChunkData(chunkPath string) ([]byte, error) {
	if !util.Exists(chunkPath) {
		return nil, fmt.Errorf("Corpus file '%s' does not exist.", chunkPath)
	}

	chunkFile, openErr := os.Open(chunkPath)
	if openErr != nil {
		return nil, openErr
	}
	defer chunkFile.Close()

	chunkStream, readErr := ioutil.ReadAll(chunkFile)
	if readErr != nil {
		return nil, fmt.Errorf("Failed to read corpus file '%s'", chunkPath)
	}

	return chunkStream, nil
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

func (chunk *Chunk) validate(reader io.Reader) bool {
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

// TODO: Put `Chunk` in its own file?
// TODO: Use validation step to populate a `Chunk.path` member?
// TODO: Rename `Chunk`. It's not a chunk, it's a raw corpus tarball.

// Chunk represents a tar'd file that contains a subset of the corpus. the
// SHA512 hash is used to verify the version of the data is correct.
type Chunk struct {
	Name   string `yaml:"name"`
	SHA512 string `yaml:"sha512"`
}
