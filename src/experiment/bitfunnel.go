package experiment

import (
	"github.com/BitFunnel/LabBook/src/bfrepo"
	"github.com/BitFunnel/LabBook/src/experiment/file"
	"github.com/BitFunnel/LabBook/src/schema"
	"github.com/BitFunnel/LabBook/src/util"
)

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

func configureBitFunnelRuntime(repo bfrepo.Manager, fileManager file.Manager, exptSchema *schema.Experiment) error {
	// Create corpus samples.
	initErr := fileManager.InitSampleCache(
		exptSchema.Samples,
		func(
			sample *schema.Sample,
			corpusManifestPath string,
			outputPath string,
		) error {
			filterErr := repo.RunFilter(
				corpusManifestPath,
				outputPath,
				sample.AsFilterArg())
			if filterErr != nil {
				return filterErr
			}
			return nil
		})
	if initErr != nil {
		return initErr
	}

	// Generate corpus statistics.
	configErr := fileManager.InitConfigCache(
		exptSchema.StatisticsConfig.SampleName,
		func(configRoot string, statsManifestPath string) error {
			statisticsErr := repo.RunStatistics(
				statsManifestPath,
				configRoot)
			if statisticsErr != nil {
				return statisticsErr
			}

			// Generate term table.
			termTableErr := repo.RunTermTable(configRoot)
			if termTableErr != nil {
				return termTableErr
			}

			return nil
		})
	if configErr != nil {
		return configErr
	}

	return nil
}
