package experiment

import (
	"io"

	"errors"

	"fmt"

	"github.com/BitFunnel/LabBook/src/bfrepo"
	"github.com/BitFunnel/LabBook/src/corpus"
	"github.com/BitFunnel/LabBook/src/experiment/file"
	"github.com/BitFunnel/LabBook/src/schema"
	"github.com/BitFunnel/LabBook/src/util"
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

	// Uncompress corpus, find filepaths of all corpus files.
	corpusManager := corpus.NewManager(schema.Corpus, expt.corpusRoot)
	uncompressErr := corpusManager.Uncompress()
	if uncompressErr != nil {
		return uncompressErr
	}
	corpusPaths, walkErr := corpusManager.GetAllCorpusFilepaths()
	if walkErr != nil {
		return fmt.Errorf("Failed to fetch filepaths of corpus rooted at "+
			"'%s':\n%v", expt.corpusRoot, walkErr)
	}

	// Write the configuration manifest we'll pass to `termtable` and
	// `statistics`. Fetch query log and write the experiment script.
	fileManager := file.NewManager(
		expt.corpusRoot,
		expt.experimentRoot,
		sampleNames(schema.Samples))
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
		schema.QueryLog.SHA512)
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

// Run deserializes an experiment schema from `reader` and attempts to run the
// experiment.
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

func configureBitFunnelRuntime(repo bfrepo.Manager, fileManager file.Manager, schema *schema.Experiment) error {
	// Create corpus samples.
	sampleDirErr := fileManager.CreateSampleDirectories()
	if sampleDirErr != nil {
		return sampleDirErr
	}
	for _, sample := range schema.Samples {
		samplePath, ok := fileManager.GetSamplePath(sample.Name)
		if !ok {
			return fmt.Errorf("Tried to create corpus sample for name '%s', "+
				"but this name didn't appear in experiment schema",
				sample.Name)
		}
		filterErr := repo.RunFilter(
			fileManager.GetConfigManifestPath(),
			samplePath,
			sample.AsFilterArg())
		if filterErr != nil {
			return filterErr
		}
	}

	// Generate corpus statistics.
	statsManifestPath, ok :=
		fileManager.GetSampleManifestPath(schema.StatisticsConfig.SampleName)
	if !ok {
		return fmt.Errorf("Statistics configuration requires sample with "+
			"name '%s', but a sample with that name was not found in "+
			"experiment schema", schema.StatisticsConfig.SampleName)
	}
	statisticsErr := repo.RunStatistics(
		statsManifestPath,
		fileManager.GetConfigRoot())
	if statisticsErr != nil {
		return statisticsErr
	}

	// Generate term table.
	termTableErr := repo.RunTermTable(fileManager.GetConfigRoot())
	if termTableErr != nil {
		return termTableErr
	}

	return nil
}

func buildBitFunnelAtRevision(bf bfrepo.Manager, revisionSha string) error {
	// Either clone or fetch the canonical BitFunnel repository.
	if !util.Exists(bf.GetPath()) {
		cloneErr := bf.Clone()
		if cloneErr != nil {
			return cloneErr
		}
	} else {
		fetchErr := bf.Fetch()
		if fetchErr != nil {
			return fetchErr
		}
	}

	// Checkout a revision related to some experiment; the `defer` will reset
	// the HEAD to what it was before we checked it out.
	checkoutHandle, checkoutErr := bf.Checkout(revisionSha)
	if checkoutErr != nil {
		return checkoutErr
	}
	defer checkoutHandle.Dispose()

	configureErr := bf.ConfigureBuild()
	if configureErr != nil {
		return configureErr
	}

	buildErr := bf.Build()
	if buildErr != nil {
		return buildErr
	}

	return nil
}

func sampleNames(samples []*schema.Sample) []string {
	var names = make([]string, len(samples))
	for index, sample := range samples {
		names[index] = sample.Name
	}

	return names
}
