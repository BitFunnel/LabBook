package experiment

import (
	"errors"
	"fmt"
	"io"

	"github.com/BitFunnel/LabBook/src/bfrepo"
	"github.com/BitFunnel/LabBook/src/corpus"
	"github.com/BitFunnel/LabBook/src/experiment/file"
	"github.com/BitFunnel/LabBook/src/schema"
)

// Experiment manages the lifecycle of the experiment, including a build of
// BitFunnel, related data, and metadata.
type Experiment interface {
	Configure(reader io.Reader) error
	Run() error
}

type experimentContext struct {
	experimentRoot string
	corpusRoot     string
	codeRepo       bfrepo.Manager
	configured     bool
	config         configContext
}

type configContext struct {
	schema        schema.Experiment
	corpusManager corpus.Manager
	fileManager   file.Manager
}

// New creates an Experiment object, which manages the lifecycle and
// resources of an experiment.
func New(experimentRoot string, bitFunnelRoot string, corpusRoot string) Experiment {
	bf := bfrepo.New(bitFunnelRoot)
	return &experimentContext{
		experimentRoot: experimentRoot,
		corpusRoot:     corpusRoot,
		codeRepo:       bf,
	}
}

// Configure completely configures the experiment that `expt` is managing. Each
// `Experiment` typically manages exactly one experiment described by an
// experiment schema YAML file, and `Configure` will do everything needed to
// configure this experiment to run. This includes everything from cloning,
// configuring, and building BitFunnel, to managing uncompressing the corpus,
// to running `statistics` and `termtable` on the corpus to configure the
// BitFunnel runtime.
func (expt *experimentContext) Configure(reader io.Reader) error {
	if expt.configured == true {
		return errors.New("Experiments can't be configured twice")
	}

	// Get YAML experiment schema describing our experiment.
	schema, deserializeError := schema.DeserializeExperimentSchema(reader)
	if deserializeError != nil {
		return deserializeError
	}

	// Check out BitFunnel at a particular commit and build.
	bfError := buildBitFunnelAtRevision(expt.codeRepo, schema.BitFunnelSha)
	if bfError != nil {
		return bfError
	}

	fileManager := file.NewManager(
		expt.corpusRoot,
		expt.experimentRoot,
		sampleNames(schema.Samples))
	corpusManager := corpus.NewManager(schema.Corpus, expt.corpusRoot)

	// Decompress corpus, find filepaths of all corpus files.
	corpusCacheErr :=
		fileManager.InitDecompressedCorpusCache(corpusManager.Decompress)
	if corpusCacheErr != nil {
		return corpusCacheErr
	}
	corpusPaths, walkErr := corpusManager.GetAllCorpusFilepaths()
	if walkErr != nil {
		return fmt.Errorf("Failed to fetch filepaths of corpus rooted at "+
			"'%s':\n%v", expt.corpusRoot, walkErr)
	}

	// Write the configuration manifest we'll pass to `termtable` and
	// `statistics`. Fetch query log and write the experiment script.
	configManifestWriteErr := fileManager.WriteConfigManifestFile(corpusPaths)
	if configManifestWriteErr != nil {
		return configManifestWriteErr
	}

	// Build statistics and term table.
	replError := configureBitFunnelRuntime(
		expt.codeRepo,
		fileManager,
		&schema)
	if replError != nil {
		return replError
	}

	// Write experiment script.
	fetchErr := fileManager.FetchMetadataAndWriteScript(
		schema.RuntimeConfig.SampleName,
		schema.QueryLog.URL,
		schema.QueryLog.FileSignature)
	if fetchErr != nil {
		return fetchErr
	}

	expt.config = configContext{
		schema:        schema,
		corpusManager: corpusManager,
		fileManager:   fileManager,
	}
	expt.configured = true

	return nil
}

// Runs the experiment described in the experiment schema `expt` is responsible
// for managing.
func (expt *experimentContext) Run() error {
	if expt.configured == false {
		return errors.New("Can't run experiment without calling `Configure`")
	}

	verifyScriptErr := expt.codeRepo.RunRepl(
		expt.config.fileManager.GetConfigRoot(),
		expt.config.fileManager.GetScriptPath())
	if verifyScriptErr != nil {
		return verifyScriptErr
	}

	return nil
}

//
// PRIVATE FUNCTIONS.
//

func decompressCorpus() error {
	return nil
}

func sampleNames(samples []*schema.Sample) []string {
	var names = make([]string, len(samples))
	for index, sample := range samples {
		names[index] = sample.Name
	}

	return names
}
