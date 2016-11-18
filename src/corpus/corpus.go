package corpus

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/BitFunnel/LabBook/src/signature"
	"github.com/BitFunnel/LabBook/src/systems/shell"
)

// Manager is responsible for the lifecycle of the corpus, including
// downloading, verifying, and uncompressing.
type Manager interface {
	Decompress() (signature signature.Signature, decompressErr error)
	GetAllCorpusFilepaths() ([]string, error)
}

type corpusContext struct {
	archive      []*ArchiveFile
	corpusRoot   string
	decompressed bool
}

// NewManager makes a `Manager`, which can be used to govern the lifecycle of a
// corpus directory.
func NewManager(archive []*ArchiveFile, corpusRoot string) Manager {
	return &corpusContext{
		archive:      archive,
		corpusRoot:   corpusRoot,
		decompressed: false,
	}
}

// Decompress will decompress the corpus files that `ctx` is responsible for
// managing.
func (ctx *corpusContext) Decompress() (signature.Signature, error) {
	if ctx.decompressed {
		return "", fmt.Errorf("Corpus at '%s' has already been uncompressed",
			ctx.corpusRoot)
	}

	// TODO: Implement locking protocol:
	// * Acquire lock. When you're finished uncompressing, write new lock with
	//   the signature.

	signatureAccumulator := signature.NewAccumulator()

	for _, archiveFile := range ctx.archive {
		archiveFilePath := ctx.getArchiveFilePath(archiveFile)

		tarballData, readErr := getArchiveFileData(archiveFilePath)
		if readErr != nil {
			return "", readErr
		}

		tarballSignature, sigErr := signatureAccumulator.AddData(tarballData)
		if sigErr != nil {
			return "", sigErr
		} else if tarballSignature != archiveFile.SHA512 {
			return "", fmt.Errorf("Signature for corpus file '%s' does not "+
				"match the hash specified in experiment YAML; it is "+
				"possible you have specified an incorrect corpus file",
				archiveFilePath)
		}

		// TODO: Probably we can avoid un-taring this all the time, and also
		// use a pure solution.
		tarErr := shell.RunCommand(
			"tar",
			"-xf",
			archiveFilePath,
			"-C",
			ctx.corpusRoot)
		if tarErr != nil {
			return "", tarErr
		}
	}

	corpusSignature, sigErr := signatureAccumulator.AccumulatedSignature()
	if sigErr != nil {
		return "", sigErr
	}

	ctx.decompressed = true

	return corpusSignature, nil
}

// GetAllCorpusFilepaths returns the absolute path of every file in the corpus.
func (ctx *corpusContext) GetAllCorpusFilepaths() ([]string, error) {
	// TODO: Consider making chunk files have a `.chunk` suffix to simplify
	// this.
	if !ctx.decompressed {
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
	// generated when we called `Decompress`). IMPORTANT: Any file we find in
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
func (ctx *corpusContext) getArchiveFilePath(archiveFile *ArchiveFile) string {
	return filepath.Join(ctx.corpusRoot, archiveFile.Name)
}
