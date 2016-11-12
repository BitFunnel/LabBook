package bfrepo

import (
	"fmt"
	"os"
	"testing"

	"github.com/BitFunnel/LabBook/src/labtest"
	"github.com/BitFunnel/LabBook/src/systems"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// LabBookTest represents a basic test suite to be run on LabBook.
type LabBookTest struct {
	suite.Suite
}

// SetupTest sets up the LabBook test suite for a new test case.
func (suite *LabBookTest) SetupTest() {
	systems.OpLog().ResetEventLog()
}

func Test_SimpleClone(t *testing.T) {
	repo := New(os.DevNull)
	cloneErr := repo.Clone()
	assert.NoError(t, cloneErr)

	assert.Equal(t, os.DevNull, repo.GetGitManager().GetRepoRootPath())

	// Verify.
	cloneCmd := fmt.Sprintf(
		`[SHELL] git clone %s %s`,
		bitfunnelHTTPSRemote,
		os.DevNull)

	eventLog := systems.OpLog().GetEventLog()
	targetLog := []string{
		cloneCmd,
	}
	labtest.AssertEventsEqual(t, targetLog, eventLog)
}

func (suite *LabBookTest) Test_FetchCheckout() {
	repo := New(".")
	repo.GetGitManager().ConfigureAsMock(
		map[string]string{"remote.origin.url": bitfunnelHTTPSRemote},
		map[string]string{"HEAD": "963091ed535b827bcbab1c607658a974679633b2"},
		map[string]string{"HEAD": "master"},
	)

	revisionSha := "4da26d9a2bf29a1eac78fb165ddb5a79caeedfb9"

	// Operations.
	{
		fetchErr := repo.Fetch()
		assert.NoError(suite.T(), fetchErr)

		checkoutHandle, checkoutErr := repo.Checkout(revisionSha)
		assert.NoError(suite.T(), checkoutErr)
		defer checkoutHandle.Dispose()
	}

	// Verify.
	wd, wdErr := os.Getwd()
	assert.NoError(suite.T(), wdErr)
	chdirCmd := fmt.Sprintf(`[FS] os.Chdir("%s")`, wd)

	checkoutCmd := fmt.Sprintf(`[SHELL] git checkout %s`, revisionSha)

	eventLog := systems.OpLog().GetEventLog()
	targetLog := []string{
		`[FS] os.Chdir(".")`,
		`[SHELL] git fetch origin`,
		chdirCmd,
		`[FS] os.Chdir(".")`,
		checkoutCmd,
		chdirCmd,
	}
	labtest.AssertEventsEqual(suite.T(), targetLog, eventLog)
}

func TestExampleTestSuite(t *testing.T) {
	suite.Run(t, new(LabBookTest))
}

func TestMain(m *testing.M) {
	systems.ConfigureAsTestRun()
	os.Exit(m.Run())
}
