package test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"
	"vcblobstore"
	"vcblobstore/git"
	"vcblobstore/git/gitlab"

	"github.com/stretchr/testify/suite"
)

type TestBlobstoreClientFactory func() (TestBlobstoreClient, error)

type BlobstoreRepository interface {
	fmt.Stringer
	CreateRepository(ctx context.Context) error
	GetBlob(ctx context.Context, key string) ([]byte, error)
	AddBlob(ctx context.Context, blob vcblobstore.BlobInfo) error
	DeleteBlob(ctx context.Context, key string, modifiedBy string) error
}

type vcblobstoreManagement interface {
	ResetRepository(ctx context.Context) error
	DeleteRepository(ctx context.Context) error
	CheckStatus() (bool, error)
	GetStateID(ctx context.Context) (string, error)
	GetVersionFor(ctx context.Context, key string) (string, error)
	GetVersionMetadata(ctx context.Context, commitId string) (git.CommitMetadata, error)
}

type TestBlobstoreClient interface {
	BlobstoreRepository
	vcblobstoreManagement
}

type TestBlobstoreController struct {
	repoFactory TestBlobstoreClientFactory
	repo        TestBlobstoreClient
}

func (ctl *TestBlobstoreController) String() string {
	return ctl.repo.String()
}

// After adding support for distributed git-repo access, "remains consistent after failed <operation>"
// should mean that the operation can be successfully repeated after an initial failure.

type BlobstoreTestSuite struct {
	suite.Suite
	RepoController TestBlobstoreController
	TestSequenceId string
	TestCaseId     int
	Ctx            context.Context
}

func TestGitRepositoryTestSuite(t *testing.T) {
	for _, repoController := range BlobstoreProvidersToTest() {
		suite.Run(t, &BlobstoreTestSuite{RepoController: repoController, TestSequenceId: "vcblobstore", Ctx: context.Background()})
	}
}

func (s *BlobstoreTestSuite) BeforeTest(suiteName, testName string) {
	var createRepoErr error
	s.RepoController.repo, createRepoErr = s.RepoController.repoFactory()
	if createRepoErr != nil {
		s.FailNow("", "%v", createRepoErr)
	}
	resetRepoErr := s.RepoController.repo.ResetRepository(s.Ctx)
	if resetRepoErr != nil {
		s.FailNow("", "%v", resetRepoErr)
	}
}

func (testSuite *gitlabRepoTestSuite) AfterTest(suiteName string, testName string) {
	testSuite.gitRepo.DeleteRepository(testSuite.ctx)
}

func (s *BlobstoreTestSuite) AssertBlobstoreCleanStatus() {
	status, err := s.RepoController.repo.CheckStatus()
	s.NoError(err)
	s.True(status)
}

func (s *BlobstoreTestSuite) AssertFileInBlobstore(blob vcblobstore.BlobInfo, timeBeforeCommit time.Time) {
	commitID, getCommitIDErr := s.RepoController.repo.GetVersionFor(s.Ctx, blob.Key)
	s.NoError(getCommitIDErr)
	s.Greater(len(commitID), 0)
	meta, getMetaErr := s.RepoController.repo.GetVersionMetadata(s.Ctx, commitID)
	s.NoError(getMetaErr)
	s.Greater(meta.CommitDate, timeBeforeCommit.Add(-time.Duration(1_000)*time.Millisecond))
}

func (s *BlobstoreTestSuite) TestAcceptsNewBlobWhenEmpty() {
	var err error
	blob := TestData[0]
	timeBeforeAdd := time.Now()
	err = s.RepoController.repo.AddBlob(s.Ctx, blob)
	s.NoError(err)

	var sha1 string
	sha1, err = s.RepoController.repo.GetStateID(s.Ctx)
	s.NoError(err)
	s.Equal(len("8e9b80b5155dea01e5175bc819bbe364dbc07a66"), len(sha1))

	s.AssertBlobstoreCleanStatus()
	s.AssertFileInBlobstore(blob, timeBeforeAdd)
}

func (s *BlobstoreTestSuite) TestAcceptsNewBlobWhenNotEmpty() {
	blob1 := TestData[0]
	blob2 := TestData[1]

	errorWhenAddingFirstBlob := s.RepoController.repo.AddBlob(s.Ctx, blob1)
	s.NoError(errorWhenAddingFirstBlob)

	firstSha1, errorWhenGettingFirstSha1 := s.RepoController.repo.GetStateID(s.Ctx)
	s.NoError(errorWhenGettingFirstSha1)
	errorAddingSecondBlob := s.RepoController.repo.AddBlob(s.Ctx, blob2)
	s.NoError(errorAddingSecondBlob)
	secondSha1, errorWhenGettingSecondSha1 := s.RepoController.repo.GetStateID(s.Ctx)
	s.NoError(errorWhenGettingSecondSha1)
	s.NotEqual(firstSha1, secondSha1)
}

func (s *BlobstoreTestSuite) TestRemainsConsistentAfterUpdatingBlobFails() {
}

func (s *BlobstoreTestSuite) TestRemainsConsistentAfterDeletingBlobFails() {
}

var DefaultBlobstoreController = TestBlobstoreController{
	repoFactory: func() (TestBlobstoreClient, error) {
		return NewLocalGitTestRepo(localTestConfig)
	},
}

var gitlabTestConfig gitlab.Config

func BlobstoreProvidersToTest() []TestBlobstoreController {
	fmt.Printf(">>>>>>>>>>>> LOCAL_GIT_ONLY: %v\n", os.Getenv("LOCAL_GIT_ONLY"))
	if len(os.Getenv("LOCAL_GIT_ONLY")) > 0 {
		return []TestBlobstoreController{DefaultBlobstoreController}
	}

	return []TestBlobstoreController{
		DefaultBlobstoreController,
		{
			repoFactory: func() (TestBlobstoreClient, error) {
				repo, createClientErr := NewGitlabTestRepoClient(&gitlabTestConfig)
				if createClientErr != nil {
					return nil, createClientErr
				}
				return repo, nil
			},
		},
	}
}
