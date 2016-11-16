package experiment

import (
	"fmt"

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

func createSample(sample *schema.Sample, repo bfrepo.Manager, fileManager file.Manager) (string, error) {
	// NOTE: The BitFunnel binary requires you to create the sample directories
	// before calling the `filter` operation. This is handled for us in the
	// `file.Manager.CacheSamples` method, but it is worth noting anyway.

	samplePath, ok := fileManager.GetSamplePath(sample.Name)
	if !ok {
		return "", fmt.Errorf("Tried to create corpus sample for name '%s', "+
			"but this name didn't appear in experiment schema",
			sample.Name)
	}
	filterErr := repo.RunFilter(
		fileManager.GetConfigManifestPath(),
		samplePath,
		sample.AsFilterArg())
	if filterErr != nil {
		return "", filterErr
	}

	// TODO: Parse the resulting manifest file for sample.
	return "", nil
}

// QUESTIONS:
// At this point there are some lingering questions, like: what does the
// codepath look like for just running an experiment? Do we need to have a
// duplicate function like `cacheSamples` for caching the configuration and so
// on? Can the steps of caching samples and statistics be done in sequence,
// (i.e., we lock samples, and then release that, and then lock again to
// generate the config details)? Are we ready to start implementing?
//
// Modulo small issues I think that `cacheSamples` demonstrates we're ready to
// start implementing. Each caching method will be pretty self-contained, able
// to operate independently of others. The execution path of different steps
// can be totally decoupled: you can cache the sample and then cache the config,
// or do them both one after another, and there will be no conflicts. There is
// some lingering issue in the execution -- when we've updated the lock files,
// for samples, then the config lock files will be out of date. How do we
// reconcile this? One idea is to just say: if you update from one step, you
// have to update everything downstream of it. So updating samples updates the
// config and experiments too. Updating the config only updates the experiment.
// And so on. Probably those lock signatures should go in the YAML when you
// finally lock an experiment, too. Also, we should think about the fact that
// we're checking out a specific version of BitFunnel every time. Maybe that's
// not desirable? We want people to be able to pick up anything that's in
// master, but also we want anyone to be able to experiment. So maybe the right
// workflow is to allow development, but then not allow you to lock the
// experiment YAML until something is written to mainline BitFunnel master?

func configureBitFunnelRuntime(repo bfrepo.Manager, fileManager file.Manager, schema *schema.Experiment) error {
	// Create corpus samples.

	// TODO: We need to un-comment this if it's going to work.
	// sampleDirErr := fileManager.CreateSampleDirectories()
	// if sampleDirErr != nil {
	// 	return sampleDirErr
	// }
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
