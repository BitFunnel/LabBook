package bfrepo

import (
	"fmt"
	"os"

	"path/filepath"

	"github.com/bitfunnel/LabBook/src/cmd"
)

const bitfunnelHTTPSRemote = `https://github.com/bitfunnel/bitfunnel`
const bitfunnelSSHRemote = `git@github.com:BitFunnel/BitFunnel.git`

// Manager manages the lifecycle of a BitFunnel repository, everything from
// cloning, to checking out a specific version, to building BitFunnel, to
// runinng the REPL.
type Manager interface {
	Path() string
	Clone() error
	Fetch() error
	Checkout(revision string) (cmd.CmdHandle, error)
	ConfigureBuild() error
	Build() error
	ConfigureRuntime(manifestFile string, configDir string) error
	Repl(configDir string, scriptFile string) error
}

type bfRepoContext struct {
	bitFunnelRoot       string
	buildRoot           string
	bitFunnelExecutable string
}

// New creates a BfRepo object, to manage a BitFunnel repository.
func New(bitFunnelRoot string) Manager {
	buildRoot := filepath.Join(bitFunnelRoot, "build-make")
	bitFunnelExecutable :=
		filepath.Join(buildRoot, "tools", "BitFunnel", "src", "BitFunnel")
	return bfRepoContext{
		bitFunnelRoot:       bitFunnelRoot,
		buildRoot:           buildRoot,
		bitFunnelExecutable: bitFunnelExecutable,
	}
}

func (repo bfRepoContext) Path() string {
	return repo.bitFunnelRoot
}

// Clone clones the canonical GitHub repository, into the folder
// `bitFunnelRoot`.
func (repo bfRepoContext) Clone() (cloneErr error) {
	cloneErr =
		cmd.RunCommand("git", "clone", bitfunnelHTTPSRemote, repo.bitFunnelRoot)
	return
}

// Fetch pulls the BitFunnel master from the canonical repository.
func (repo bfRepoContext) Fetch() error {
	chdirHandle, chdirErr := scopedChdir(repo.bitFunnelRoot)
	if chdirErr != nil {
		return chdirErr
	}
	defer chdirHandle.Dispose()

	originURL, originURLErr :=
		cmd.CommandOutput("git", "config", "--get", "remote.origin.url")
	if originURLErr != nil {
		return originURLErr
	}

	if originURL != bitfunnelSSHRemote && originURL != bitfunnelHTTPSRemote {
		return fmt.Errorf("The remote 'origin' in the repository located at "+
			"%s' is required to point at the canonical BitFunnel repository.",
			repo.bitFunnelRoot)
	}

	pullErr := cmd.RunCommand("git", "fetch", "origin")
	if pullErr != nil {
		return pullErr
	}
	return nil
}

// Checkout take a path to a canonical BitFunnel repository,
// `bitFunnelRoot`, and checks out a commit from the canonical GitHub
// repository, specified by `sha`.
func (repo bfRepoContext) Checkout(sha string) (cmd.CmdHandle, error) {
	chdirHandle, chdirErr := scopedChdir(repo.bitFunnelRoot)
	if chdirErr != nil {
		return nil, chdirErr
	}
	defer chdirHandle.Dispose()

	// Returns the "short name" of HEAD. Usually this is a branch, like
	// `master`, but if HEAD is detached, it can also simply be `HEAD`.
	headRef, headRefErr :=
		cmd.CommandOutput("git", "rev-parse", "--abbrev-ref=strict", "HEAD")
	if headRefErr != nil {
		return nil, headRefErr
	}

	// The commit hash for HEAD.
	headSha, headShaErr := cmd.CommandOutput("git", "rev-parse", "HEAD")
	if headShaErr != nil {
		return nil, headShaErr
	}

	// Checkout commit denoted with `sha`.
	checkoutErr := cmd.RunCommand("git", "checkout", sha)
	if checkoutErr != nil {
		return nil, checkoutErr
	}

	// Set dispose to reset the head when we're done with it.
	resetHead := func() error {
		chdirHandle, chdirErr := scopedChdir(repo.bitFunnelRoot)
		if chdirErr != nil {
			return chdirErr
		}
		defer chdirHandle.Dispose()

		var presentRef string
		if headRef == "HEAD" {
			presentRef = headSha
		} else {
			presentRef = headRef
		}

		checkoutErr := cmd.RunCommand("git", "checkout", presentRef)
		return checkoutErr
	}

	return cmd.MakeHandle(resetHead), nil
}

// Configure switches to the directory of the BitFunnel root, and runs
// the configuration script that generates a makefile.
func (repo bfRepoContext) ConfigureBuild() error {
	chdirHandle, chdirErr := scopedChdir(repo.bitFunnelRoot)
	if chdirErr != nil {
		return chdirErr
	}
	defer chdirHandle.Dispose()

	configErr := cmd.RunCommand("sh", "Configure_Make.sh")
	return configErr
}

// Build switches to the BitFunnel build directory, and builds the code.
func (repo bfRepoContext) Build() error {
	chdirHandle, chdirErr := scopedChdir(repo.buildRoot)
	if chdirErr != nil {
		return chdirErr
	}
	defer chdirHandle.Dispose()

	buildErr := cmd.RunCommand("make", "-j4")
	return buildErr
}

// Repl runs the BitFunnel repl.
func (repo bfRepoContext) ConfigureRuntime(manifestFile string, configDir string) error {
	return cmd.RunCommand(
		repo.bitFunnelExecutable,
		"statistics",
		manifestFile,
		configDir,
		"-text")
}

// Repl runs the BitFunnel repl.
func (repo bfRepoContext) Repl(configDir string, scriptFile string) error {
	return cmd.RunCommand(
		repo.bitFunnelExecutable,
		"repl",
		configDir,
		"-script",
		scriptFile)
}

// scopedChdir changes to `directory` and then, when `Dispose` is
// called, it changes back to the current working directory.
func scopedChdir(directory string) (cmd.CmdHandle, error) {
	pwd, pwdErr := os.Getwd()
	if pwdErr != nil {
		return nil, pwdErr
	}

	chdirErr := os.Chdir(directory)
	if chdirErr != nil {
		return nil, chdirErr
	}

	return cmd.MakeHandle(func() error { return os.Chdir(pwd) }), nil
}
