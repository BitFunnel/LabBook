package cli

import (
	"flag"

	"github.com/BitFunnel/LabBook/src/cli/errors"
	"github.com/BitFunnel/LabBook/src/systems"
)

const labUsage = `Usage:
  lab [-simulate] [-verbose] run <experiment-yaml> <bitfunnelRoot> <experimentRoot> <corpusRoot>
`

// ParseAndDispatch parses and dispatches the command for the `lab` binary.
func ParseAndDispatch(arguments []string) {
	if len(arguments) <= 1 {
		errors.Logger.Fatal(labUsage)
	}

	userSafetyErr := checkCurrentUserSafe()
	errors.CheckFatalB(userSafetyErr)

	dryRun := flag.Bool(
		"simulate",
		false,
		"Print potentially-deleterious operations instead of performing them.")
	verbose := flag.Bool(
		"verbose",
		false,
		"Pipe all output to stdout, including output for: configuration, build, git operations, etc.")
	flag.Parse()

	if *dryRun {
		systems.ConfigureAsDryRun()
	}

	if *verbose {
		systems.ConfigureAsVerboseRun()
	}

	if flag.Args()[0] == "run" {
		labRun(flag.Args())
	} else {
		errors.Logger.Fatalf("Invalid command '%s'.\n%s", arguments[1], labRunUsage)
	}
}
