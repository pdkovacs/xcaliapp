package s3store

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"
)

type s3SessionStoreTest struct {
	suite.Suite
	ctx   context.Context
	store *SessionStore
}

func TestXCaliSessionStoreS3(t *testing.T) {
	suite.Run(t, &s3SessionStoreTest{})
}

func (s *s3SessionStoreTest) SetupSuite() {
	fmt.Println("SetupSuite running...")
	s.ctx = context.Background()
	var err error
	s.store, err = NewSessionStore(s.ctx, "test-xcalidrawings")
	s.NoError(err)
}

func (s *s3SessionStoreTest) TestGetAllowedCredentials() {
	creds, getCredsErr := s.store.GetAllowedCredentials(s.ctx)
	s.NoError(getCredsErr)
	s.Equal("qwer\n", creds)
}

func (s *s3SessionStoreTest) TestCreateListSessions() {
	_, createSessErr := s.store.CreateSession(s.ctx)
	s.NoError(createSessErr)
	sessId1, createSessErr1 := s.store.CreateSession(s.ctx)
	s.NoError(createSessErr1)

	sessionList, listErr := s.store.ListSessions(s.ctx)
	s.NoError(listErr)

	s.Equal(1, len(sessionList))
	s.Contains(sessionList, sessId1)
}
