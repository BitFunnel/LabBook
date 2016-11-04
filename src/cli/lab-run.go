package cli

import (
	"os"
	"path/filepath"

	"github.com/bitfunnel/LabBook/src/cli/errors"
	"github.com/bitfunnel/LabBook/src/experiment"
)

const labRunUsage = `Usage:
  lab run <experiment-yaml> <bitfunnelRoot> <experimentRoot> <corpusRoot>
`

func labRun(arguments []string) {
	// Validate arguments.
	if len(arguments) != 5 {
		errors.Logger.Fatalf("Invalid number of arguments.\n%s", labRunUsage)
	}

	schemaPath, absErr := filepath.Abs(arguments[1])
	errors.CheckFatal(absErr, "Schema path is not well-formed")
	bitFunnelRoot, absErr := filepath.Abs(arguments[2])
	errors.CheckFatal(absErr, "BitFunnel root path is not well-formed")
	experimentRoot, absErr := filepath.Abs(arguments[3])
	errors.CheckFatal(absErr, "Experiment path is not well-formed")
	corpusRoot, absErr := filepath.Abs(arguments[4])
	errors.CheckFatal(absErr, "Corpus root path is not well-formed")

	// Check experiment directory is safe to write to.
	safetyErr := checkDirectorySafe(experimentRoot)
	errors.CheckFatalB(safetyErr)

	// Get schema file.
	schemaFile, fErr := os.Open(schemaPath)
	errors.CheckFatalB(fErr)
	defer schemaFile.Close()

	// Configure and run experiment.
	expt := experiment.New(
		experimentRoot,
		bitFunnelRoot,
		corpusRoot)
	configErr := expt.Configure(schemaFile)
	errors.CheckFatal(configErr, "Failed to configure experiment")

	runErr := expt.Run()
	errors.CheckFatal(runErr, "Failed to run experiment")
}
