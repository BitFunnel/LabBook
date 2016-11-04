package filesystem

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
)

// Manager manages the lifecycle of the experiment files. This includes
// fetching remote script and manifest files, writing them to the config, and
// so on.
type Manager interface {
	ScriptPath() string
	FetchAndWriteMetadata() error
}

// New creates a new Manager object.
func New(configRoot string, corpusRoot string, manifestURL string, scriptURL string) Manager {
	scriptPath := fmt.Sprintf("%s/script.txt", configRoot)
	return managerContext{
		configRoot:  configRoot,
		corpusRoot:  corpusRoot,
		manifestURL: manifestURL,
		scriptURL:   scriptURL,
		scriptPath:  scriptPath,
	}
}

type managerContext struct {
	configRoot  string
	corpusRoot  string
	manifestURL string
	scriptURL   string
	scriptPath  string
}

func (m managerContext) FetchAndWriteMetadata() error {
	paths, fetchManifestErr := fetchFileLines(m.manifestURL)
	if fetchManifestErr != nil {
		return fetchManifestErr
	}

	script, fetchScriptErr := fetchFileLines(m.scriptURL)
	if fetchScriptErr != nil {
		return fetchScriptErr
	}

	return m.writeScript(paths, script)
}

func (m managerContext) ScriptPath() string {
	return m.scriptPath
}

func (m managerContext) writeScript(manifestPaths []string, script []string) error {
	mkdirErr := os.MkdirAll(m.configRoot, 0777)
	if mkdirErr != nil {
		return mkdirErr
	}

	w, createErr := os.Create(m.scriptPath)
	if createErr != nil {
		return createErr
	}
	defer w.Close()

	for _, path := range manifestPaths {
		if path == "" {
			continue
		}
		w.WriteString(fmt.Sprintf("cache chunk %s/%s\n", m.corpusRoot, path))
	}

	for _, line := range script {
		w.WriteString(line)
	}

	w.Sync()

	return nil
}

// TODO: Make this an actual URL.
func fetchFileLines(url string) ([]string, error) {
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

	lines := strings.Split(string(file), "\n")

	// writeErr := ioutil.WriteFile(path, file, 0777)
	// if writeErr != nil {
	// 	return "", writeErr
	// }

	return lines, nil
}
