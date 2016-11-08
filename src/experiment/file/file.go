package file

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/bitfunnel/LabBook/src/systems/fs"
	"github.com/bitfunnel/LabBook/src/util"
)

// Manager manages the lifecycle of the experiment files. This includes
// fetching remote script and manifest files, writing them to the config, and
// so on.
type Manager interface {
	ScriptPath() string
	ConfigManifestPath() string
	WriteConfigManifestFile(absoluteCorpusPaths []string) error
	FetchMetadataAndWriteScript(queryLogURL *url.URL, queryLogSHA512 string) error
}

// NewManager creates a new Manager object.
func NewManager(configRoot string, corpusRoot string, experimentRoot string) Manager {
	configManifestPath := filepath.Join(configRoot, "config_manifest.txt")
	scriptPath := filepath.Join(configRoot, "script.txt")
	verifyOutPath := filepath.Join(experimentRoot, "verify_out")
	noVerifyOutPath := filepath.Join(experimentRoot, "no_verify_out")
	return managerContext{
		configRoot:         configRoot,
		corpusRoot:         corpusRoot,
		scriptPath:         scriptPath,
		verifyOutPath:      verifyOutPath,
		noVerifyOutPath:    noVerifyOutPath,
		configManifestPath: configManifestPath,
	}
}

type managerContext struct {
	configRoot         string
	corpusRoot         string
	scriptPath         string
	verifyOutPath      string
	noVerifyOutPath    string
	configManifestPath string
}

// WriteConfigManifestFile takes a list of absolute paths to corpus files, and
// writes them to the manifest file.
func (m managerContext) WriteConfigManifestFile(absoluteCorpusPaths []string) error {
	fileBytes := []byte(strings.Join(absoluteCorpusPaths, "\n"))
	writeErr := fs.WriteFile(m.configManifestPath, fileBytes, 0666)
	if writeErr != nil {
		return fmt.Errorf("Failed to write configuration manifest file at "+
			"'%s':\n%v", m.configManifestPath, writeErr)
	}

	return nil
}

func (m managerContext) FetchMetadataAndWriteScript(queryLogURL *url.URL, queryLogSHA512 string) error {
	queryLog, queryLogFetchErr := fetchFileLines(
		queryLogURL.String(),
		queryLogSHA512)
	if queryLogFetchErr != nil {
		return queryLogFetchErr
	}

	ingestManifestLines, readErr := readFileLines(m.configManifestPath)
	if readErr != nil {
		return fmt.Errorf("Failed to read ingestion manifest at '%s' for "+
			"experiment:\n%v", m.configManifestPath, readErr)
	}

	writeScriptErr := m.writeScript(ingestManifestLines, queryLog)
	if writeScriptErr != nil {
		return fmt.Errorf("Failed to write verifying script at path "+
			"'%s':\n%v", m.scriptPath, writeScriptErr)
	}

	return nil
}

func (m managerContext) ScriptPath() string {
	return m.scriptPath
}

func (m managerContext) ConfigManifestPath() string {
	return m.configManifestPath
}

func (m managerContext) writeScript(manifestPaths []string, queryLog []string) error {
	mkdirErr := fs.MkdirAll(m.configRoot, 0777)
	if mkdirErr != nil {
		return mkdirErr
	}
	mkdirErr = fs.MkdirAll(m.verifyOutPath, 0777)
	if mkdirErr != nil {
		return mkdirErr
	}
	mkdirErr = fs.MkdirAll(m.noVerifyOutPath, 0777)
	if mkdirErr != nil {
		return mkdirErr
	}

	// TODO: Check to see if this will overwrite, rather than append, if it
	// already exists.
	w, createErr := fs.Create(m.scriptPath)
	if createErr != nil {
		return createErr
	}
	defer w.Close()

	for _, path := range manifestPaths {
		if path == "" {
			continue
		}
		_, writeErr := w.WriteString(
			fmt.Sprintf("cache chunk %s/%s\n", m.corpusRoot, path))
		if writeErr != nil {
			return fmt.Errorf("Failed to write script file at '%s':\n%v",
				m.scriptPath, writeErr)
		}
	}

	writeErr := m.writeQueriesToScript(w, queryLog, true)
	if writeErr != nil {
		return fmt.Errorf("Failed to write script file at '%s':\n%v",
			m.scriptPath, writeErr)
	}

	m.writeQueriesToScript(w, queryLog, false)
	if writeErr != nil {
		return fmt.Errorf("Failed to write script file at '%s':\n%v",
			m.scriptPath, writeErr)
	}

	syncErr := w.Sync()
	if syncErr != nil {
		return fmt.Errorf("Failed to write script file at '%s':\n%v",
			m.scriptPath, syncErr)
	}

	return nil
}

func (m managerContext) writeQueriesToScript(w *os.File, queryLog []string, verify bool) error {
	var outPath string
	var queryBasis string
	if verify {
		outPath = m.verifyOutPath
		queryBasis = "verify one %s\n"
	} else {
		outPath = m.noVerifyOutPath
		queryBasis = "query one %s\n"
	}

	_, writeErr := w.WriteString(fmt.Sprintf("cd %s\n", outPath))
	if writeErr != nil {
		return writeErr
	}

	for _, query := range queryLog {
		if query == "" {
			continue
		}

		_, writeErr = w.WriteString(fmt.Sprintf(queryBasis, query))
		if writeErr != nil {
			return writeErr
		}
	}

	_, writeErr = w.WriteString("analyze")
	if writeErr != nil {
		return writeErr
	}

	return nil
}

// TODO: Make this an actual URL.
func fetchFileLines(url string, validationSHA512 string) ([]string, error) {
	resp, getErr := http.Get(url)
	if getErr != nil {
		return nil, getErr
	} else if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Failed to get file at '%s'; returned code '%d'", url, resp.StatusCode)
	}
	defer resp.Body.Close()

	file, readErr := ioutil.ReadAll(resp.Body)
	if readErr != nil {
		return nil, readErr
	}

	if !util.ValidateSHA512(file, validationSHA512) {
		return nil, fmt.Errorf("File located at resource '%s' does not "+
			"match given SHA512 '%s'. Does this point to the version of the "+
			"file required by the experiment?", url, validationSHA512)
	}

	lines := strings.Split(string(file), "\n")

	return lines, nil
}

func readFileLines(path string) ([]string, error) {
	file, openErr := os.Open(path)
	if openErr != nil {
		return nil, openErr
	}
	defer file.Close()

	fileBytes, readErr := ioutil.ReadAll(file)
	if readErr != nil {
		return nil, readErr
	}

	return strings.Split(string(fileBytes), "\n"), nil
}
