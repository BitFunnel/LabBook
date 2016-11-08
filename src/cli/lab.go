package cli

import (
	"flag"

	"github.com/BitFunnel/LabBook/src/cli/errors"
	"github.com/BitFunnel/LabBook/src/systems"
)

const labUsage = `Usage:
  lab [-dry-run] run <experiment-yaml> <bitfunnelRoot> <experimentRoot> <corpusRoot>
`

// ParseAndDispatch parses and dispatches the command for the `lab` binary.
func ParseAndDispatch(arguments []string) {
	if len(arguments) <= 1 {
		errors.Logger.Fatal(labUsage)
	}

	userSafetyErr := checkCurrentUserSafe()
	errors.CheckFatalB(userSafetyErr)

	// TODO: This flag parsing is awful. We can't even allow -dry-run to appear
	// last. We should replace this ASAP.
	dryRun := flag.Bool(
		"dry-run",
		false,
		"Print potentially-deleterious operations instead of performing them.")
	flag.Parse()

	if *dryRun {
		systems.ConfigureAsDryRun()
	}

	if flag.Args()[0] == "run" {
		labRun(flag.Args())
	} else {
		errors.Logger.Fatalf("Invalid command '%s'.\n%s", arguments[1], labRunUsage)
	}
}
