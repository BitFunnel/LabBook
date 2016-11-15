package file

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/BitFunnel/LabBook/src/systems/fs"
	"github.com/BitFunnel/LabBook/src/util"
)

// Manager manages the lifecycle of the experiment files. This includes
// fetching remote script and manifest files, writing them to the config, and
// so on.
type Manager interface {
	CreateSampleDirectories() error
	WriteConfigManifestFile(absoluteCorpusPaths []string) error
	FetchMetadataAndWriteScript(sampleName string, queryLogURL *url.URL, queryLogSHA512 string) error
	GetConfigRoot() string
	GetConfigManifestPath() string
	GetSamplePath(sampleName string) (string, bool)
	GetSampleManifestPath(sampleName string) (string, bool)
	GetScriptPath() string
}

// NewManager creates a new Manager object.
func NewManager(corpusRoot string, experimentRoot string, sampleNames []string) Manager {
	configRoot := filepath.Join(experimentRoot, "configuration")
	configManifestPath := filepath.Join(configRoot, "config_manifest.txt")
	runtimeManifestPath := filepath.Join(configRoot, "runtime_manifest.txt")
	scriptPath := filepath.Join(configRoot, "script.txt")
	verifyOutPath := filepath.Join(experimentRoot, "verify_out")
	noVerifyOutPath := filepath.Join(experimentRoot, "no_verify_out")
	sampleRoot := filepath.Join(experimentRoot, "samples")

	samplePaths := make(map[string]string, len(sampleNames))
	for _, sampleName := range sampleNames {
		samplePaths[sampleName] = filepath.Join(sampleRoot, sampleName)
	}

	return managerContext{
		configRoot:          configRoot,
		corpusRoot:          corpusRoot,
		sampleRoot:          sampleRoot,
		samplePaths:         samplePaths,
		scriptPath:          scriptPath,
		verifyOutPath:       verifyOutPath,
		noVerifyOutPath:     noVerifyOutPath,
		configManifestPath:  configManifestPath,
		runtimeManifestPath: runtimeManifestPath,
	}
}

type managerContext struct {
	configRoot          string
	corpusRoot          string
	sampleRoot          string
	samplePaths         map[string]string
	scriptPath          string
	verifyOutPath       string
	noVerifyOutPath     string
	configManifestPath  string
	runtimeManifestPath string
}

// CreateSampleDirectories will create the directories we'll need to generate
// the filtered samples for an experiment. For example, if an experiment
// defines 2 samples with names `sample1` and `sample2`, this will create
// directories for each.
func (m managerContext) CreateSampleDirectories() error {
	for _, samplePath := range m.samplePaths {
		mkdirFilteredCorpusRoot := fs.MkdirAll(samplePath, 0777)
		if mkdirFilteredCorpusRoot != nil {
			return fmt.Errorf("Unable to create filtered corpus directory "+
				"'%s':\n%v", samplePath, mkdirFilteredCorpusRoot)
		}
	}

	return nil
}

// WriteConfigManifestFile takes a list of absolute paths to corpus files, and
// writes them to the manifest file.
func (m managerContext) WriteConfigManifestFile(absoluteCorpusPaths []string) error {
	mkConfigRootErr := fs.MkdirAll(m.configRoot, 0777)
	if mkConfigRootErr != nil {
		return mkConfigRootErr
	}

	fileBytes := []byte(strings.Join(absoluteCorpusPaths, "\n"))
	writeErr := fs.WriteFile(m.configManifestPath, fileBytes, 0666)
	if writeErr != nil {
		return fmt.Errorf("Failed to write configuration manifest file at "+
			"'%s':\n%v", m.configManifestPath, writeErr)
	}

	return nil
}

// FetchMetadataAndWriteScript will fetch the metadata needed to generate a
// script for an experiment (e.g., a query log), and then generates and writes
// the script.
func (m managerContext) FetchMetadataAndWriteScript(sampleName string, queryLogURL *url.URL, queryLogSHA512 string) error {
	queryLog, queryLogFetchErr := fetchFileLines(
		queryLogURL.String(),
		queryLogSHA512)
	if queryLogFetchErr != nil {
		return queryLogFetchErr
	}

	runtimeManifestFile, ok := m.GetSampleManifestPath(sampleName)
	ingestManifestLines, readErr := readFileLines(runtimeManifestFile)
	if !ok || readErr != nil {
		return fmt.Errorf("Failed to read ingestion manifest at '%s' for "+
			"experiment:\n%v", runtimeManifestFile, readErr)
	}

	writeScriptErr := m.writeScript(ingestManifestLines, queryLog)
	if writeScriptErr != nil {
		return fmt.Errorf("Failed to write verifying script at path "+
			"'%s':\n%v", m.scriptPath, writeScriptErr)
	}

	return nil
}

// GetConfigRoot will get the path to the root of the directory that contains
// configuration information for BitFunnel's runtime (e.g., files for the term
// table, statistics, etc.)
func (m managerContext) GetConfigRoot() string {
	return m.configRoot
}

// GetConfigManifestPath will return the path to the canonical manifest file of
// the corpus sample used to configure the experiment. Typically this is given
// to the BitFunnel tools to generate a sample of the corpus to runtime
// `statistics` and `termtable` on (for example).
func (m managerContext) GetConfigManifestPath() string {
	return m.configManifestPath
}

// GetSamplePath will get the canonical directory meant to hold the sample of
// the corpus denoted by `sampleName`. For example, if we have a sample called
// `sample1`, this will return the directory that corresponds to that sample.
func (m managerContext) GetSamplePath(sampleName string) (string, bool) {
	manifestPath, ok := m.samplePaths[sampleName]
	return manifestPath, ok
}

// GetSampleManifestPath will return the canonical manifest file generated for
// an experiment's corpus sample. For example, if we have a sample named
// `sample1`, this will return a manifest file listing all the files in the
// corpus sample `sample1`.
func (m managerContext) GetSampleManifestPath(sampleName string) (string, bool) {
	manifestPath, ok := m.samplePaths[sampleName]
	return filepath.Join(manifestPath, "Manifest.txt"), ok
}

// GetScriptPath will return the path to the canonical script file for the
// experiment.
func (m managerContext) GetScriptPath() string {
	// TODO: Check that this was actually generated?
	return m.scriptPath
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
			fmt.Sprintf("cache chunk %s\n", path))
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

	_, writeErr = w.WriteString("analyze\n")
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
	file, openErr := fs.Open(path)
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
