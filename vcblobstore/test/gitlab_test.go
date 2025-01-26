package test

import (
	"vcblobstore/git/gitlab"
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/suite"
)

type gitlabRepoTestSuite struct {
	suite.Suite
	ctx     context.Context
	t       *testing.T
	gitRepo *gitlab.Gitlab
}

func TestGitlabRepoTestSuite(t *testing.T) {
	if len(os.Getenv("LOCAL_GIT_ONLY")) > 0 {
		return
	}
	suite.Run(t, &gitlabRepoTestSuite{ctx: context.Background(), t: t})
}

func (testSuite *gitlabRepoTestSuite) TestAddIconfile() {
	var err error
	blob := TestData[0]
	err = testSuite.gitRepo.AddBlob(testSuite.ctx, blob)
	testSuite.NoError(err)

	// var sha1 string
	// sha1, err = testSuite.GetStateID()
	testSuite.NoError(err)
	// testSuite.Equal(len("8e9b80b5155dea01e5175bc819bbe364dbc07a66"), len(sha1))
	// testSuite.assertGitCleanStatus()
	// testSuite.assertFileInRepo(icon.Name, iconfile)
}
