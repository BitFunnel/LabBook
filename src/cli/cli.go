package cli

import (
	"log"
	"os"
	"path/filepath"

	"github.com/bitfunnel/bf-lab-notebook/src/experiment"
	"github.com/bitfunnel/bf-lab-notebook/src/util"
)

const usage = `
Usage:
  bflab run experiment.yaml <bitfunnelRoot> <experimentRoot> <corpusRoot>
`

// TODO: Rename this.
func Process(arguments []string) {
	if len(arguments) != 6 {
		log.Fatalf("Invalid number of arguments.\n%s", usage)
	} else if arguments[1] != "run" {
		log.Fatalf("Invalid command '%s'.\n%s", arguments[1], usage)
	}

	// TODO: Source this to the CLI logic.
	// TODO: Have this return a path instead.
	schemaPath, absErr := filepath.Abs(arguments[2])
	util.Check(absErr, "Schema path is not well-formed")
	experimentRoot, absErr := filepath.Abs(arguments[4])
	util.Check(absErr, "Experiment path is not well-formed")
	bitFunnelRoot, absErr := filepath.Abs(arguments[3])
	util.Check(absErr, "BitFunnel root path is not well-formed")
	corpusRoot, absErr := filepath.Abs(arguments[5])
	util.Check(absErr, "Corpus root path is not well-formed")

	safetyErr := safetyCheck(experimentRoot)
	util.CheckErr(safetyErr)

	schemaFile, fErr := os.Open(schemaPath)
	util.CheckErr(fErr)

	expt := experiment.New(
		experimentRoot,
		bitFunnelRoot,
		corpusRoot)
	configErr := expt.Configure(schemaFile)
	util.Check(configErr, "Failed to configure experiment")

	runErr := expt.Run()
	util.Check(runErr, "Failed to run experiment")
}
