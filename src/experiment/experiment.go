package experiment

import (
	"fmt"
	"io"

	"errors"

	"github.com/bitfunnel/LabBook/src/bfrepo"
	"github.com/bitfunnel/LabBook/src/corpus"
	"github.com/bitfunnel/LabBook/src/experiment/filesystem"
	"github.com/bitfunnel/LabBook/src/util"
)

// Experiment manages the lifecycle of the experiment, including a build of
// BitFunnel, related data, and metadata.
type Experiment interface {
	Configure(reader io.Reader) error
	Run() error
}

type experimentContext struct {
	experimentRoot string
	configRoot     string
	corpusRoot     string
	codeRepo       bfrepo.Manager
	configured     bool
	config         configContext
}

type configContext struct {
	schema      Schema
	corpus      corpus.Manager
	fileManager filesystem.Manager
}

// New creates an Experiment object, which manages the lifecycle and
// resources of an experiment.
func New(experimentRoot string, bitFunnelRoot string, corpusRoot string) Experiment {
	configRoot := fmt.Sprintf("%s/configuration", experimentRoot)
	bf := bfrepo.New(bitFunnelRoot)
	return &experimentContext{
		experimentRoot: experimentRoot,
		configRoot:     configRoot,
		corpusRoot:     corpusRoot,
		codeRepo:       bf,
	}
}

func (expt *experimentContext) Configure(reader io.Reader) error {
	schema, deserializeError := DeserializeSchema(reader)
	if deserializeError != nil {
		return deserializeError
	}

	bfError := buildBitFunnelAtRevision(expt.codeRepo, schema.BfSha)
	if bfError != nil {
		return bfError
	}

	corpus := corpus.New(schema.Corpus, expt.corpusRoot)
	uncompressErr := corpus.Uncompress()
	if uncompressErr != nil {
		return uncompressErr
	}

	filesystem := filesystem.New(
		expt.configRoot,
		expt.corpusRoot,
		schema.ManifestFile,
		schema.ScriptFile)
	fetchErr := filesystem.FetchAndWriteMetadata()
	if fetchErr != nil {
		return fetchErr
	}

	// replError := expt.codeRepo.ConfigureRuntime(
	// 	"/Users/alex/src/BitFunnel/wikidata/enwiki-20161020-config")
	// if replError != nil {
	// 	return replError
	// }

	expt.config = configContext{
		schema:      schema,
		corpus:      corpus,
		fileManager: filesystem,
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

	replError := expt.codeRepo.Repl(
		"/Users/alex/src/BitFunnel/wikidata/enwiki-20161020-config",
		expt.config.fileManager.ScriptPath())
	if replError != nil {
		return replError
	}

	return nil
}

func buildBitFunnelAtRevision(bf bfrepo.Manager, revisionSha string) error {
	// Either clone or fetch the canonical BitFunnel repository.
	if !util.Exists(bf.Path()) {
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
