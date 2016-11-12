package experiment

import (
	"fmt"

	"github.com/BitFunnel/LabBook/src/bfrepo"
	"github.com/BitFunnel/LabBook/src/experiment/file"
	"github.com/BitFunnel/LabBook/src/schema"
	"github.com/BitFunnel/LabBook/src/util"
)

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
	if !util.Exists(bf.GetGitManager().GetRepoRootPath()) {
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
