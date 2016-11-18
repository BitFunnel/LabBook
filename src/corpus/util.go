package corpus

import (
	"fmt"
	"os"

	"github.com/BitFunnel/LabBook/src/systems/mockablefs"
	"github.com/BitFunnel/LabBook/src/util"
)

func getArchiveFileData(archivePath string) (archiveData []byte, err error) {
	if !util.Exists(archivePath) {
		return nil, fmt.Errorf("Corpus file '%s' does not exist.", archivePath)
	}

	openErr := mockablefs.OpenDo(
		archivePath,
		func(archiveFileData []byte) error {
			archiveData = archiveFileData
			return nil
		})
	if openErr != nil {
		return nil, fmt.Errorf("Failed to read corpus file '%s'", archivePath)
	}

	return
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
