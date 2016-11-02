package main

import (
	"log"
	"os"

	"github.com/bitfunnel/bf-lab-notebook/src/experiment"
	"github.com/bitfunnel/bf-lab-notebook/src/util"
)

const usage = `
Usage:
  bflab run experiment.yaml <bitfunnelRoot> <experimentRoot> <corpusRoot>
`

func main() {
	if len(os.Args) != 6 {
		log.Fatalf("Invalid number of arguments.\n%s", usage)
	} else if os.Args[1] != "run" {
		log.Fatalf("Invalid command '%s'.\n%s", os.Args[1], usage)
	}

	schemaFile, fErr := os.Open(os.Args[2])
	if fErr != nil {
		log.Fatal(fErr)
	}

	expt := experiment.New(
		os.Args[4],
		os.Args[3],
		os.Args[5])
	configErr := expt.Configure(schemaFile)
	util.Check(configErr, "Failed to configure experiment")

	runErr := expt.Run()
	util.Check(runErr, "Failed to run experiment")
}
