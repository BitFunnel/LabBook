package shell

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/BitFunnel/LabBook/src/systems"
)

type shellOperation struct {
	command string
	args    []string
}

func (shellOp *shellOperation) String() string {
	// TODO: Probably we need to check the size of tab to make this perfect
	// everywhere.
	return fmt.Sprintf("[SHELL]\t\t%s %s", shellOp.command, strings.Join(shellOp.args, " "))
}

func newShellOperation(command string, args []string) systems.Operation {
	return &shellOperation{command: command, args: args}
}

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
	if systems.IsDryRun() {
		systems.OpLog().Log(newShellOperation(command, args))
		return nil
	}

	cmd := exec.Command(command, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// CommandOutput synchronously executes a command, captures the stdout, and
// returns it as a string.
func CommandOutput(command string, args ...string) (string, error) {
	if systems.IsDryRun() {
		systems.OpLog().Log(newShellOperation(command, args))

		if systems.IsTestRun() {
			return "", nil
		}
	}

	cmd := exec.Command(command, args...)
	cmd.Stderr = os.Stderr
	output, cmdErr := cmd.Output()
	return strings.TrimSpace(string(output)), cmdErr
}

// MakeHandle makes a command handle that calls `cleanup` during a `Dispose`.
func MakeHandle(cleanup func() error) CmdHandle {
	return &scopedCommand{dispose: cleanup}
}

type scopedCommand struct {
	dispose func() error
}

// Dispose performs the cleanup operation for a `CmdHandle`. For example, if
// we've run `os.Chdir` and returned a `CmdHandle`, we might have `Dispose`
// call `os.Chdir` to return to the original directory we were in.
func (c *scopedCommand) Dispose() error {
	return c.dispose()
}
