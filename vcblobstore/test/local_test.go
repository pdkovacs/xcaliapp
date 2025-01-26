package test

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
	"vcblobstore/git"
	"vcblobstore/git/local"

	"github.com/stretchr/testify/suite"
)

const testRepoLocation = "test-git-repo"

var localTestConfig = &local.Config{
	Location: filepath.Join(os.Getenv("HOME"), "tmp", "versioned-vcblobstore", testRepoLocation),
}

var localGitRepoTestLogger = createTestLogger()

type localGitRepoTestSuite struct {
	suite.Suite
	t             *testing.T
	gitRepoClient *local.Git
	ctx           context.Context
}

func TestLocalGitTestSuite(t *testing.T) {
	repo, createRepoErr := NewLocalGitTestRepo(localTestConfig)
	if createRepoErr != nil {
		panic(createRepoErr)
	}
	suite.Run(t, &localGitRepoTestSuite{t: t, gitRepoClient: repo})
}

func (testSuite *localGitRepoTestSuite) removeRepoDir() {
	rmdirErr := os.RemoveAll(testRepoLocation)
	if rmdirErr != nil {
		panic(rmdirErr)
	}
}

func (testSuite *localGitRepoTestSuite) BeforeTest(suiteName string, testName string) {
	logger := localGitRepoTestLogger.With().Str("root", "TestLocalGitTestSuite").Logger()
	testSuite.ctx = logger.WithContext(context.Background())
	testSuite.removeRepoDir()
	gitRepoCreationError := testSuite.gitRepoClient.CreateRepository(testSuite.ctx)
	if gitRepoCreationError != nil {
		panic(gitRepoCreationError)
	}
}

func (testSuite *localGitRepoTestSuite) TestLocationHasRepo() {
	err := testSuite.gitRepoClient.CreateRepository(testSuite.ctx)
	if err != nil {
		panic(err)
	}
	err = testSuite.gitRepoClient.DeleteRepository(testSuite.ctx)
	if err != nil {
		panic(err)
	}
	err = testSuite.gitRepoClient.CreateRepository(testSuite.ctx)
	if err != nil {
		panic(err)
	}
	testSuite.Equal(true, testSuite.gitRepoClient.LocationHasRepo())
}

func (testSuite *localGitRepoTestSuite) TestParseCommitMetadata() {
	// git show --quiet --format=fuller --date=format:'%Y-%m-%dT%H:%M:%S%z' 95fb34a325e697bafffb785ac65ecca986ca06a6

	testInput := `Author:     Kovács, Péter <peter.dunay.kovacs@gmail.com>
AuthorDate: 2022-10-09T13:42:12+0200
Commit:     Péter Kovács <peter.dunay.kovacs@gmail.com>
CommitDate: 2022-10-31T15:30:17+0100
	
    [DEV] distributed git access`

	authorDate, authorDateErr := time.Parse(time.RFC3339, "2022-10-09T13:42:12+02:00")
	testSuite.Nil(authorDateErr)
	commitDate, commitDateErr := time.Parse(time.RFC3339, "2022-10-31T15:30:17+01:00")
	testSuite.Nil(commitDateErr)

	expectedOutput := git.CommitMetadata{
		Author:     "Kovács, Péter <peter.dunay.kovacs@gmail.com>",
		AuthorDate: authorDate,
		Commit:     "Péter Kovács <peter.dunay.kovacs@gmail.com>",
		CommitDate: commitDate,
		Message:    "[DEV] distributed git access",
	}

	commitMetadata, parseErr := git.ParseLocalCommitMetadata(testInput)
	testSuite.Nil(parseErr)
	testSuite.Equal(expectedOutput, commitMetadata)
}

func NewLocalGitTestRepo(conf *local.Config) (*local.Git, error) {
	testLogger := createTestLogger()
	repo := local.NewLocalGitRepository(conf, &testLogger)
	return repo, nil
}
