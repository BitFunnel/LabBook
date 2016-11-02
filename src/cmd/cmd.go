package cmd

import (
	"os"
	"os/exec"
	"strings"
)

// CmdHandle allows automatic cleanup after complex commands. For example,
// if we want to `chdir` to one directory, and then at the end of scope,
// `chdir` back to the present working directory, we would set up a command
// handle that, when we call `Dispose`, changes back to the original working
// directory.
type CmdHandle interface {
	Dispose() error
}

// RunCommand synchronously executes a command and pipes the output to stderr
// and stdout.
func RunCommand(command string, args ...string) error {
	cmd := exec.Command(command, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// CommandOutput synchronously executes a command, captures the stdout, and
// returns it as a string.
func CommandOutput(command string, args ...string) (string, error) {
	cmd := exec.Command(command, args...)
	cmd.Stderr = os.Stderr
	output, cmdErr := cmd.Output()
	return strings.TrimSpace(string(output)), cmdErr
}

// MakeHandle makes a command handle that calls `cleanup` during a `Dispose`.
func MakeHandle(cleanup func() error) CmdHandle {
	return scopedCommand{dispose: cleanup}
}

type scopedCommand struct {
	dispose func() error
}

func (c scopedCommand) Dispose() error {
	return c.dispose()
}
