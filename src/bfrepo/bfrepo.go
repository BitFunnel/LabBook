package bfrepo

import (
	"fmt"
	"os"

	"path/filepath"

	"strings"

	"github.com/BitFunnel/LabBook/src/systems/fs"
	"github.com/BitFunnel/LabBook/src/systems/shell"
)

// NOTE: Git remotes are case-insensitive, which is why they're lowercase here.
const bitfunnelHTTPSRemote = `https://github.com/bitfunnel/bitfunnel`
const bitfunnelSSHRemote = `git@github.com:bitfunnel/bitfunnel.git`

// Manager manages the lifecycle of a BitFunnel repository, everything from
// cloning, to checking out a specific version, to building BitFunnel, to
// runinng the REPL.
type Manager interface {
	Path() string
	Clone() error
	Fetch() error
	Checkout(revision string) (shell.CmdHandle, error)
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
		shell.RunCommand("git", "clone", bitfunnelHTTPSRemote, repo.bitFunnelRoot)
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
		shell.CommandOutput("git", "config", "--get", "remote.origin.url")
	if originURLErr != nil {
		return originURLErr
	}

	lowerOriginURL := strings.ToLower(originURL)

	if lowerOriginURL != bitfunnelSSHRemote &&
		lowerOriginURL != bitfunnelHTTPSRemote {
		return fmt.Errorf("The remote 'origin' in the repository located at "+
			"%s' is required to point at the canonical BitFunnel repository.",
			repo.bitFunnelRoot)
	}

	pullErr := shell.RunCommand("git", "fetch", "origin")
	if pullErr != nil {
		return pullErr
	}
	return nil
}

// Checkout take a path to a canonical BitFunnel repository,
// `bitFunnelRoot`, and checks out a commit from the canonical GitHub
// repository, specified by `sha`.
func (repo bfRepoContext) Checkout(sha string) (shell.CmdHandle, error) {
	chdirHandle, chdirErr := scopedChdir(repo.bitFunnelRoot)
	if chdirErr != nil {
		return nil, chdirErr
	}
	defer chdirHandle.Dispose()

	// Returns the "short name" of HEAD. Usually this is a branch, like
	// `master`, but if HEAD is detached, it can also simply be `HEAD`.
	headRef, headRefErr :=
		shell.CommandOutput("git", "rev-parse", "--abbrev-ref=strict", "HEAD")
	if headRefErr != nil {
		return nil, headRefErr
	}

	// The commit hash for HEAD.
	headSha, headShaErr := shell.CommandOutput("git", "rev-parse", "HEAD")
	if headShaErr != nil {
		return nil, headShaErr
	}

	// Checkout commit denoted with `sha`.
	checkoutErr := shell.RunCommand("git", "checkout", sha)
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

		checkoutErr := shell.RunCommand("git", "checkout", presentRef)
		return checkoutErr
	}

	return shell.MakeHandle(resetHead), nil
}

// Configure switches to the directory of the BitFunnel root, and runs
// the configuration script that generates a makefile.
func (repo bfRepoContext) ConfigureBuild() error {
	chdirHandle, chdirErr := scopedChdir(repo.bitFunnelRoot)
	if chdirErr != nil {
		return chdirErr
	}
	defer chdirHandle.Dispose()

	configErr := shell.RunCommand("sh", "Configure_Make.sh")
	return configErr
}

// Build switches to the BitFunnel build directory, and builds the code.
func (repo bfRepoContext) Build() error {
	chdirHandle, chdirErr := scopedChdir(repo.buildRoot)
	if chdirErr != nil {
		return chdirErr
	}
	defer chdirHandle.Dispose()

	buildErr := shell.RunCommand("make", "-j4")
	return buildErr
}

// Repl runs the BitFunnel repl.
func (repo bfRepoContext) ConfigureRuntime(manifestFile string, configDir string) error {
	// TODO: Filter corpus here also.

	statisticsErr := shell.RunCommand(
		repo.bitFunnelExecutable,
		"statistics",
		manifestFile,
		configDir,
		"-text")
	if statisticsErr != nil {
		return statisticsErr
	}

	termTableErr := shell.RunCommand(
		repo.bitFunnelExecutable,
		"termtable",
		configDir)
	if termTableErr != nil {
		return termTableErr
	}

	return nil
}

// Repl runs the BitFunnel repl.
func (repo bfRepoContext) Repl(configDir string, scriptFile string) error {
	return shell.RunCommand(
		repo.bitFunnelExecutable,
		"repl",
		configDir,
		"-script",
		scriptFile)
}

// scopedChdir changes to `directory` and then, when `Dispose` is
// called, it changes back to the current working directory.
func scopedChdir(directory string) (shell.CmdHandle, error) {
	pwd, pwdErr := os.Getwd()
	if pwdErr != nil {
		return nil, pwdErr
	}

	chdirErr := fs.Chdir(directory)
	if chdirErr != nil {
		return nil, chdirErr
	}

	return shell.MakeHandle(func() error { return fs.Chdir(pwd) }), nil
}
