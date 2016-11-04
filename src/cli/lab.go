package cli

import "github.com/bitfunnel/LabBook/src/cli/errors"

const labUsage = `Usage:
  lab run <experiment-yaml> <bitfunnelRoot> <experimentRoot> <corpusRoot>
`

// ParseAndDispatch parses and dispatches the command for the `lab` binary.
func ParseAndDispatch(arguments []string) {
	if len(arguments) <= 1 {
		errors.Logger.Fatal(labUsage)
	}

	userSafetyErr := checkCurrentUserSafe()
	errors.CheckFatalB(userSafetyErr)

	if arguments[1] == "run" {
		labRun(arguments[1:])
	} else {
		errors.Logger.Fatalf("Invalid command '%s'.\n%s", arguments[1], labRunUsage)
	}
}
