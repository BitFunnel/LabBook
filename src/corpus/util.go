package corpus

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/BitFunnel/LabBook/src/util"
)

func getArchiveFileData(archivePath string) ([]byte, error) {
	if !util.Exists(archivePath) {
		return nil, fmt.Errorf("Corpus file '%s' does not exist.", archivePath)
	}

	archiveFile, openErr := os.Open(archivePath)
	if openErr != nil {
		return nil, openErr
	}
	defer archiveFile.Close()

	archiveFileStream, readErr := ioutil.ReadAll(archiveFile)
	if readErr != nil {
		return nil, fmt.Errorf("Failed to read corpus file '%s'", archivePath)
	}

	return archiveFileStream, nil
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
