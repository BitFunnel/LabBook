package file

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/BitFunnel/LabBook/src/experiment/file/lock"
	"github.com/BitFunnel/LabBook/src/schema"
	"github.com/BitFunnel/LabBook/src/signature"
	"github.com/BitFunnel/LabBook/src/systems/fs"
)

const lockFileName = "LOCKFILE"
const tmpLockFileName = ".LOCKFILE"

// CacheCorpusOperation represents a cachable corpus decompression operation.
// When the corpus is decompressed, this function should generate a signature
// for all the data in the corpus, and return it. We can then later use this
// signature to verify that the data in the corpus is what we think it is.
type CacheCorpusOperation func() (signature signature.Signature, decompressErr error)

// CacheSampleOperation represents the operation of generating a cachable
// sample of a corpus. The function should take a schema describing the sample
// to generate, as well as a path to the manifest and directory path for
// results, and run BitFunnel's `filter` command, returning an error as
// appropriate.
type CacheSampleOperation func(
	sample *schema.Sample,
	corpusManifestPath string,
	outputPath string,
) error

// ConfigCacheOperation represents the operation of generating a cachable
// BitFunnel runtime configuration. The function should take a directory path
// that configuration information can be dropped into, as well as a manifest
// path for data on which to run the configuration commands. An error should be
// returned as appropriate.
type ConfigCacheOperation func(
	configRoot string,
	statsManifestPath string,
) error

// Manager manages the lifecycle of the experiment files. This includes
// fetching remote script and manifest files, writing them to the config, and
// so on.
type Manager interface {
	// TODO: Consider moving these into their own module -- caching is complex
	// and we want to get it right.
	InitDecompressedCorpusCache(decompressCorpus CacheCorpusOperation) error
	UpdateDecompressedCorpusCache(decompressCorpus CacheCorpusOperation) error

	InitSampleCache(samples []*schema.Sample, createSample CacheSampleOperation) error
	UpdateSampleCache(samples []*schema.Sample, createSample CacheSampleOperation) error
	VerifySampleCache() error

	InitConfigCache(sampleName string, configure ConfigCacheOperation) error

	WriteConfigManifestFile(absoluteCorpusPaths []string) error
	// TODO: Move the "fetch" part of this out of `file.Manager`. This should
	// only manage files.
	FetchMetadataAndWriteScript(sampleName string, queryLogURL *url.URL, queryLogSHA512 signature.Signature) error

	// Methods that return paths we can send to BitFunnel's shell commands.
	GetConfigRoot() string
	GetConfigManifestPath() string
	GetSamplePath(sampleName string) (string, bool)
	GetSampleManifestPath(sampleName string) (string, bool)
	GetScriptPath() string
}

// NewManager creates a new Manager object.
func NewManager(corpusRoot string, experimentRoot string, sampleNames []string) Manager {
	configRoot := filepath.Join(experimentRoot, "configuration")
	sampleRoot := filepath.Join(experimentRoot, "samples")

	samplePaths := make(map[string]string, len(sampleNames))
	for _, sampleName := range sampleNames {
		samplePaths[sampleName] = filepath.Join(sampleRoot, sampleName)
	}

	return &managerContext{
		experimentRoot:      experimentRoot,
		configRoot:          configRoot,
		corpusRoot:          corpusRoot,
		sampleRoot:          sampleRoot,
		samplePaths:         samplePaths,
		scriptPath:          filepath.Join(configRoot, "script.txt"),
		verifyOutPath:       filepath.Join(experimentRoot, "verify_out"),
		noVerifyOutPath:     filepath.Join(experimentRoot, "no_verify_out"),
		configManifestPath:  filepath.Join(configRoot, "config_manifest.txt"),
		runtimeManifestPath: filepath.Join(configRoot, "runtime_manifest.txt"),
	}
}

type managerContext struct {
	experimentRoot      string
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

// CacheDecompressedCorpus initializes a corpus cache: it will decompress a
// corpus, cache it, and generate a lock file for the cache. Note that, because
// this is an initialization procedure, we do not acquire the lock file, hence,
// we are not protected against multiple processes.
func (m *managerContext) InitDecompressedCorpusCache(decompressCorpus CacheCorpusOperation) error {
	corpusSignature, decompressErr := decompressCorpus()
	if decompressErr != nil {
		return decompressErr
	}

	lockPath := filepath.Join(m.experimentRoot, lockFileName)
	corpusLock := lock.NewCorpusLockFile(lockPath, corpusSignature)

	writeErr := m.writeLockFile(lockPath, corpusLock)
	if writeErr != nil {
		return writeErr
	}

	return nil
}

func (m *managerContext) UpdateDecompressedCorpusCache(
	decompressCorpus CacheCorpusOperation,
) error {
	corpusLock, lockAcqErr := m.acquireLockFile(m.experimentRoot)
	if lockAcqErr != nil {
		return lockAcqErr
	}

	corpusSignature, decompressErr := decompressCorpus()
	if decompressErr != nil {
		return decompressErr
	}

	corpusLock.UpdateSignature(corpusSignature)

	// Release the lock file if and only if we're successful with the above.
	return m.releaseLockFile(m.experimentRoot, corpusLock)
}

func (m *managerContext) VerifySampleCache() error {
	corpusLock, lockAcqErr := m.acquireLockFile(m.experimentRoot)
	if lockAcqErr != nil {
		return lockAcqErr
	}
	defer m.releaseLockFile(m.experimentRoot, corpusLock)

	sampleLock, lockAcqErr := m.acquireLockFile(m.sampleRoot)
	if lockAcqErr != nil {
		return lockAcqErr
	}
	defer m.releaseLockFile(m.sampleRoot, sampleLock)

	return lock.ValidateSampleLockFile(corpusLock, sampleLock)
}

func (m *managerContext) InitSampleCache(
	samples []*schema.Sample,
	createSample CacheSampleOperation,
) error {
	// TODO: Consider erroring out if we haven't called the corpus init
	// routines.

	corpusLock, lockAcqErr := m.acquireLockFile(m.experimentRoot)
	if lockAcqErr != nil {
		return lockAcqErr
	}
	defer m.releaseLockFile(m.experimentRoot, corpusLock)

	// The BitFunnel executable does not create directories as part of its
	// `filter` operation, so we create sample directoreis on its behalf.
	sampleDirErr := m.createSampleDirectories()
	if sampleDirErr != nil {
		return sampleDirErr
	}

	// Create corpus samples, write a lock file for each.
	for _, sample := range samples {
		samplePath, ok := m.GetSamplePath(sample.Name)
		if !ok {
			return fmt.Errorf("Tried to create corpus sample for name '%s', "+
				"but this name didn't appear in experiment schema",
				sample.Name)
		}

		filterErr := createSample(sample, m.configManifestPath, samplePath)
		if filterErr != nil {
			return filterErr
		}

		// TODO: This is gross and could probably be merged into the
		// conditional above.
		sampleManifestPath, ok := m.GetSampleManifestPath(sample.Name)
		if !ok {
			return fmt.Errorf("Tried to create corpus sample for name '%s', "+
				"but this name didn't appear in experiment schema",
				sample.Name)
		}

		sampleFiles, readErr := readFileLines(sampleManifestPath)
		if readErr != nil {
			return readErr
		}

		sampleSignature, signatureError := m.createSignature(sampleFiles)
		if signatureError != nil {
			return signatureError
		}

		// Create and write out a lockfile for the sample.
		sampleLockPath := filepath.Join(samplePath, lockFileName)
		sampleLock := lock.NewSampleLockFile(
			sampleLockPath,
			sampleSignature,
			corpusLock.Signature())

		writeErr := m.writeLockFile(sampleLockPath, sampleLock)
		if writeErr != nil {
			return writeErr
		}
	}

	return nil
}

func (m *managerContext) InitConfigCache(sampleName string, configure ConfigCacheOperation) error {
	// TODO: Consider erroring out if we haven't called the sample init
	// routines.

	samplePath, samplePathExists := m.GetSamplePath(sampleName)
	sampleManifestPath, sampleManifestPathExists :=
		m.GetSampleManifestPath(sampleName)
	if !(samplePathExists && sampleManifestPathExists) {
		return fmt.Errorf("Statistics configuration requires sample with "+
			"name '%s', but a sample with that name was not found in "+
			"experiment schema", sampleName)
	}

	sampleLock, lockAcqErr := m.acquireLockFile(samplePath)
	if lockAcqErr != nil {
		return lockAcqErr
	}
	defer m.releaseLockFile(samplePath, sampleLock)

	configureErr := configure(
		m.configRoot,
		sampleManifestPath)
	if configureErr != nil {
		return configureErr
	}

	sampleFiles, matchErr := m.getConfigPaths()
	if matchErr != nil {
		return matchErr
	}

	configSignature, signatureError := m.createSignature(sampleFiles)
	if signatureError != nil {
		return signatureError
	}

	configLockPath := filepath.Join(m.configRoot, lockFileName)
	configLock := lock.NewConfigLockFile(
		configLockPath,
		configSignature,
		sampleLock.Signature())

	writeErr := m.writeLockFile(configLockPath, configLock)
	if writeErr != nil {
		return writeErr
	}

	return nil
}

func (m *managerContext) UpdateSampleCache(samples []*schema.Sample, createSample CacheSampleOperation) error {
	return nil
	// // `Acquire` should either delete LOCKFILE, or move LOCKFILE -> .LOCKFILE.
	// // Then deserialize, return struct here. Rationale is: because we are not
	// // changing the corpus (or the corpus lock), we will always want put it
	// // back. It's not important that it's moved to .LOCKFILE, it just seems
	// // convenient.
	// corpusLock, lockAcqErr := m.acquireLockFile(m.experimentRoot)
	// if lockAcqErr != nil {
	// 	// This error should already explain the problem, e.g., if the
	// 	// .LOCKFILE exists and LOCKFILE does not.
	// 	return lockAcqErr
	// }
	// // Put the LOCKFILE back. We didn't touch the corpus, so we always want to
	// // put it back. This also ensures that we put the corpus file back after we
	// // write out the sample lock.
	// // NOTE: We need to figure out how to deal with the `error` that
	// // `ReleaseLock` returns. Probably name the `error` return parameter, and
	// // wrap this in a func that sets it in the case of error.
	// // TODO: Look at entire codebase for bad uses of `defer`.
	// defer m.releaseLockFile(m.experimentRoot, corpusLock)

	// // TODO: Probably want to acquire the locks of all samples.

	// // Attempt to acquire the LOCKFILE for samples. As above, this will either
	// // delete LOCKFILE or move it. This time we only want to write out a new
	// // LOCKFILE on success.
	// sampleLock, lockAcqErr := m.acquireLockFile(m.sampleRoot)
	// if lockAcqErr != nil {
	// 	return lockAcqErr
	// }

	// // Create Samples.
	// sampleDirErr := m.createSampleDirectories()
	// if sampleDirErr != nil {
	// 	return sampleDirErr
	// }

	// for _, sample := range samples {
	// 	sampleFiles, filterErr := createSample(sample)
	// 	if filterErr != nil {
	// 		return filterErr
	// 	}

	// 	// TODO: Probably want to accumulate
	// 	_, signatureError := m.createSignature(sampleFiles)
	// 	if signatureError != nil {
	// 		return signatureError
	// 	}
	// }

	// // TODO: We need to set the sample signature, and release all of the locks.

	// // Write out lock only on success.
	// return m.releaseLockFile(m.sampleRoot, sampleLock)
}

// WriteConfigManifestFile takes a list of absolute paths to corpus files, and
// writes them to the manifest file.
func (m *managerContext) WriteConfigManifestFile(absoluteCorpusPaths []string) error {
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
func (m *managerContext) FetchMetadataAndWriteScript(
	sampleName string,
	queryLogURL *url.URL,
	queryLogSHA512 signature.Signature,
) error {
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

//
// PATH GETTER METHODS. These are typically passed to BitFunnel shell commands.
// TODO: MOVE THESE TO THEIR OWN FILE.
//

// GetConfigRoot will get the path to the root of the directory that contains
// configuration information for BitFunnel's runtime (e.g., files for the term
// table, statistics, etc.)
func (m *managerContext) GetConfigRoot() string {
	return m.configRoot
}

// GetConfigManifestPath will return the path to the canonical manifest file of
// the corpus sample used to configure the experiment. Typically this is given
// to the BitFunnel tools to generate a sample of the corpus to runtime
// `statistics` and `termtable` on (for example).
func (m *managerContext) GetConfigManifestPath() string {
	return m.configManifestPath
}

// GetSamplePath will get the canonical directory meant to hold the sample of
// the corpus denoted by `sampleName`. For example, if we have a sample called
// `sample1`, this will return the directory that corresponds to that sample.
func (m *managerContext) GetSamplePath(sampleName string) (string, bool) {
	manifestPath, ok := m.samplePaths[sampleName]
	return manifestPath, ok
}

// GetSampleManifestPath will return the canonical manifest file generated for
// an experiment's corpus sample. For example, if we have a sample named
// `sample1`, this will return a manifest file listing all the files in the
// corpus sample `sample1`.
func (m *managerContext) GetSampleManifestPath(sampleName string) (string, bool) {
	manifestPath, ok := m.samplePaths[sampleName]
	return filepath.Join(manifestPath, "Manifest.txt"), ok
}

// GetScriptPath will return the path to the canonical script file for the
// experiment.
func (m *managerContext) GetScriptPath() string {
	// TODO: Check that this was actually generated?
	return m.scriptPath
}

//
// PRIVATE METHODS.
//

func (m *managerContext) acquireLockFile(
	directory string,
) (lock.Manager, error) {
	lockPath := filepath.Join(directory, lockFileName)
	tmpLockPath := filepath.Join(directory, tmpLockFileName)

	// Attempt to atomically acquire the lock by creating a hard link from
	// `LOCKFILE` to `.LOCKFILE`. Only one process can create the hard link, so
	// whichever process does not error out during this call can safely
	// proceed.

	linkErr := fs.Link(lockPath, tmpLockPath)
	if linkErr != nil {
		if os.IsNotExist(linkErr) {
			return nil, sourceDoesNotExistError(lockPath)
		} else if os.IsExist(linkErr) {
			return nil, destinationExistsError(tmpLockPath)
		} else {
			return nil, unknownLockError(lockPath, linkErr)
		}
	}

	removeErr := fs.Remove(lockPath)
	if removeErr != nil {
		return nil, couldNotRemoveSourceError(lockPath, removeErr)
	}

	lockFile, readErr := m.readLockFile(tmpLockPath)
	if readErr != nil {
		return nil, readErr
	}

	return lockFile, nil
}

func (m *managerContext) releaseLockFile(
	directory string,
	lockFile lock.Manager,
) error {
	lockPath := filepath.Join(directory, lockFileName)
	tmpLockPath := filepath.Join(directory, tmpLockFileName)

	writeErr := m.writeLockFile(tmpLockPath, lockFile)
	if writeErr != nil {
		return writeErr
	}

	linkErr := fs.Link(tmpLockPath, lockPath)
	if linkErr != nil {
		if os.IsNotExist(linkErr) {
			// return sourceDoesNotExistError(lockPath)
			return fmt.Errorf("Attempted to to release lock file, but '%s' does not exist (this should not happen, please file a bug)", tmpLockPath)
		} else if os.IsExist(linkErr) {
			// return destinationExistsError(tmpLockPath)
			return fmt.Errorf("Attempted to to release lock file, but '%s' already exists (this should not happen, please file a bug)", lockPath)
		} else {
			// return unknownLockError(lockPath, linkErr)
			return fmt.Errorf("Attempted to to release lock file at '%s', but unknown error happened (this should not happen, please file a bug):\n%v", tmpLockPath, linkErr)
		}
	}

	removeErr := fs.Remove(tmpLockPath)
	if removeErr != nil {
		// return couldNotRemoveSourceError(lockPath, removeErr)
		return fmt.Errorf("Attempted to to release lock file, but we could not remove '%s' (this should not happen, please file a bug):\n%v", tmpLockPath, linkErr)
	}

	return nil
}

func (m *managerContext) createSignature(
	dataPaths []string,
) (signature.Signature, error) {
	signatureAccumulator := signature.NewAccumulator()
	for _, path := range dataPaths {
		// TODO: This will cause -dry-run to fail because we never add teh
		// empty bytes to the signature, which causes the call to signature to
		// fail.
		if path == "" {
			continue
		}

		// TODO: Probably we want to move this to some method.

		// Get content.
		file, openErr := fs.Open(path)
		if openErr != nil {
			return "", fmt.Errorf("Attempted to create signature for data, but failed to open path '%s':\n%v", path, openErr)
		}
		defer file.Close()

		fileBytes, readErr := ioutil.ReadAll(file)
		if readErr != nil {
			return "", fmt.Errorf("Attempted to create signature for data, but failed to read file at path '%s':\n%v", path, readErr)
		}

		// Add content to signature.
		_, sigErr := signatureAccumulator.AddData(fileBytes)
		if sigErr != nil {
			return "", sigErr
		}
	}

	return signatureAccumulator.AccumulatedSignature()
}

// createSampleDirectories will create the directories we'll need to generate
// the filtered samples for an experiment. For example, if an experiment
// defines 2 samples with names `sample1` and `sample2`, this will create
// directories for each.
func (m *managerContext) createSampleDirectories() error {
	for _, samplePath := range m.samplePaths {
		mkdirFilteredCorpusRoot := fs.MkdirAll(samplePath, 0777)
		if mkdirFilteredCorpusRoot != nil {
			return fmt.Errorf("Unable to create filtered corpus directory "+
				"'%s':\n%v", samplePath, mkdirFilteredCorpusRoot)
		}
	}

	return nil
}

func (m *managerContext) writeLockFile(
	lockPath string,
	lockFile lock.Manager,
) error {
	lockFileData, createErr := fs.Create(lockPath)
	if createErr != nil {
		return fmt.Errorf("Attempt to write lock file '%s' failed:\n%v", lockPath, createErr)
	}
	defer lockFileData.Close()

	serializeErr := lock.SerializeLockFile(lockFile, lockFileData)
	if serializeErr != nil {
		// TODO: Consider deleting the lockfile if we can't write to it.
		return fmt.Errorf("Attempt to serialize and write lock file '%s' failed:\n%v ", lockPath, serializeErr)
	}
	// TODO: Verify that other calls to `Create` use this as well.
	lockFileData.Sync()

	return nil
}

func (m *managerContext) readLockFile(
	lockPath string,
) (lock.Manager, error) {
	lockFileData, readErr := fs.Open(lockPath)
	if readErr != nil {
		return nil, fmt.Errorf("Attempted to read lock file '%s', but failed:\n%v", lockPath, readErr)
	}
	defer lockFileData.Close()

	lockFile, deserErr := lock.DeserializeLockFile(lockFileData, lockPath)
	if deserErr != nil {
		return nil, fmt.Errorf("Attempted to read and deserialize lock file '%s', but failed:\n%v", lockPath, deserErr)
	}

	return lockFile, nil
}

func (m *managerContext) writeScript(manifestPaths []string, queryLog []string) error {
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

func (m *managerContext) writeQueriesToScript(w *os.File, queryLog []string, verify bool) error {
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

//
// PRIVATE FUNCTIONS.
//

func (m *managerContext) getConfigPaths() ([]string, error) {
	termCountPaths, _ := filepath.Glob(
		filepath.Join(
			m.configRoot,
			"CumulativeTermCounts-*.csv"))
	freqTablePaths, _ := filepath.Glob(
		filepath.Join(
			m.configRoot,
			"DocFreqTable-*.csv"))
	indexedIdfTablePaths, _ := filepath.Glob(
		filepath.Join(
			m.configRoot,
			"IndexedIdfTable-*.bin"))
	termTablePaths, _ := filepath.Glob(
		filepath.Join(
			m.configRoot,
			"TermTable-*.bin"))
	docLengthHistPaths, _ := filepath.Glob(
		filepath.Join(
			m.configRoot,
			"DocumentLengthHistogram.csv"))
	termToTextPaths, _ := filepath.Glob(
		filepath.Join(
			m.configRoot,
			"TermToText.bin"))

	if len(termCountPaths) <= 0 || len(freqTablePaths) <= 0 || len(indexedIdfTablePaths) <= 0 || len(termTablePaths) <= 0 || len(docLengthHistPaths) <= 0 || len(termToTextPaths) <= 0 {
		return nil, fmt.Errorf("Configuration directory '%s' must contain at least one of each of the following files: cumulative term counts, document frequency table, indexed IDF table, term table, document length histogram, and term table text mapping", m.configRoot)
	}

	allPaths := []string{}
	allPaths = append(allPaths, termCountPaths...)
	allPaths = append(allPaths, freqTablePaths...)
	allPaths = append(allPaths, indexedIdfTablePaths...)
	allPaths = append(allPaths, termTablePaths...)
	allPaths = append(allPaths, docLengthHistPaths...)
	allPaths = append(allPaths, termToTextPaths...)
	return allPaths, nil
}

// TODO: Make this an actual URL.
func fetchFileLines(
	url string,
	validationSHA512 signature.Signature,
) ([]string, error) {
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

	if !signature.ValidateData(file, validationSHA512) {
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
