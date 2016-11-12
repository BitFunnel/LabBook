package git

import (
	"fmt"

	"github.com/BitFunnel/LabBook/src/systems/fs"
	"github.com/BitFunnel/LabBook/src/systems/shell"
)

const gitCommand = "git"

// RepoManager is responsible for managing a git repository, dispatching `git`
// commands to the shell.
type RepoManager interface {
	ConfigureAsMock(
		variables map[string]string,
		revParseRefs map[string]string,
		revParseStrictRefs map[string]string,
	)
	GetRepoRootPath() string
	CloneFromOrigin() error
	GetConfig(variable string) (string, error)
	Fetch(remote string) error
	GetRevParseStrictRef(ref string) (string, error)
	GetRevParseRef(ref string) (string, error)
	Checkout(sha string) error
}

// NewRepoManager creates a BfRepo object, to manage a BitFunnel repository.
func NewRepoManager(originRemoteURL string, repoRoot string) RepoManager {
	// TODO: Make `bitFunnelRoot` an absolute path, or move this to `path`.
	return &repoContext{
		repoRoot:        repoRoot,
		originRemoteURL: originRemoteURL,
		mockConfig:      MockContext{configuredAsMock: false},
	}
}

type repoContext struct {
	repoRoot        string
	originRemoteURL string
	mockConfig      MockContext
}

// MockContext maintains information that allows us to mock `git` commands.
type MockContext struct {
	configuredAsMock   bool
	variables          map[string]string
	revParseRefs       map[string]string
	revParseStrictRefs map[string]string
}

func getOrError(dict map[string]string, key string) (string, error) {
	val, ok := dict[key]
	if !ok {
		return "", fmt.Errorf("Failed to get value for key '%s'", key)
	}

	return val, nil
}

// ConfigureAsMock configures the `RepoManager` to return mocked values for
// functions that get information, such as `GetConfig`.
func (repo *repoContext) ConfigureAsMock(
	variables map[string]string,
	revParseRefs map[string]string,
	revParseStrictRefs map[string]string,
) {
	repo.mockConfig = MockContext{
		configuredAsMock:   true,
		variables:          variables,
		revParseRefs:       revParseRefs,
		revParseStrictRefs: revParseStrictRefs,
	}
}

// GetRepoRootPath returns the root of the git repository that this
// `RepoManager` manages.
func (repo *repoContext) GetRepoRootPath() string {
	return repo.repoRoot
}

// Clone runs `git clone` in a shell.
func (repo *repoContext) CloneFromOrigin() error {
	return shell.RunCommand(
		gitCommand,
		"clone",
		repo.originRemoteURL,
		repo.repoRoot)
}

// GetConfig runs the `git config --get` command that returns the value of
// some `variable`.
func (repo *repoContext) GetConfig(variable string) (string, error) {
	if repo.mockConfig.configuredAsMock {
		return getOrError(repo.mockConfig.variables, variable)
	}

	chdirHandle, chdirErr := fs.ScopedChdir(repo.repoRoot)
	if chdirErr != nil {
		return "", chdirErr
	}
	defer chdirHandle.Dispose()

	return shell.CommandOutput(gitCommand, "config", "--get", variable)
}

// Fetch runs the `git fetch` command in a shell.
func (repo *repoContext) Fetch(remote string) error {
	chdirHandle, chdirErr := fs.ScopedChdir(repo.repoRoot)
	if chdirErr != nil {
		return chdirErr
	}
	defer chdirHandle.Dispose()

	return shell.RunCommand(gitCommand, "fetch", remote)
}

// GetRevParseStrictRef runs the `git rev-parse` command that returns the "short
// name" of `ref`. For example, if `ref` is `HEAD`, it will usually return
// either the name of the branch we're on, or the commit hash if we're in a
// detached head.
func (repo *repoContext) GetRevParseStrictRef(ref string) (string, error) {
	if repo.mockConfig.configuredAsMock {
		return getOrError(repo.mockConfig.revParseStrictRefs, ref)
	}

	chdirHandle, chdirErr := fs.ScopedChdir(repo.repoRoot)
	if chdirErr != nil {
		return "", chdirErr
	}
	defer chdirHandle.Dispose()

	return shell.CommandOutput(
		gitCommand,
		"rev-parse",
		"--abbrev-ref=strict",
		ref)
}

// GetRevParseRef runs the `git rev-parse` command that returns the commit hash
// of `ref`. For example, if `ref` is `HEAD`, this will return the commit hash
// of `HEAD`.
func (repo *repoContext) GetRevParseRef(ref string) (string, error) {
	if repo.mockConfig.configuredAsMock {
		return getOrError(repo.mockConfig.revParseRefs, ref)
	}

	chdirHandle, chdirErr := fs.ScopedChdir(repo.repoRoot)
	if chdirErr != nil {
		return "", chdirErr
	}
	defer chdirHandle.Dispose()

	return shell.CommandOutput(gitCommand, "rev-parse", ref)
}

// Checkout runs the `git checkout` command in a shell.
func (repo *repoContext) Checkout(sha string) error {
	chdirHandle, chdirErr := fs.ScopedChdir(repo.repoRoot)
	if chdirErr != nil {
		return chdirErr
	}
	defer chdirHandle.Dispose()

	return shell.RunCommand(gitCommand, "checkout", sha)
}
