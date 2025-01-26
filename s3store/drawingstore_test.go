package s3store

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"
)

type s3StoreTest struct {
	suite.Suite
	ctx   context.Context
	store *DrawingStore
}

func TestXCaliDrawingStoreS3(t *testing.T) {
	suite.Run(t, &s3StoreTest{})
}

func (s *s3StoreTest) SetupSuite() {
	fmt.Println("SetupSuite running...")
	s.ctx = context.Background()
	var err error
	s.store, err = NewDrawingStore(s.ctx, "test-xcalidrawings")
	s.NoError(err)
}

func (s *s3StoreTest) TestListDrawings() {
	inputTitles := []string{
		"some title",
		"some other title",
	}
	inputContents := []string{
		"some content",
		"some other content",
	}

	for index := range inputTitles {
		contentReader := strings.NewReader(inputContents[index])
		errPut := s.store.PutDrawing(s.ctx, inputTitles[index], contentReader, "test-user")
		s.NoError(errPut)
	}

	outputTitles, getTitlesErr := s.store.ListDrawingTitles(s.ctx)
	s.NoError(getTitlesErr)
	s.Equal(2, len(outputTitles))

	slices.Sort(inputTitles)
	slices.Sort(outputTitles)
	s.Equal(inputTitles, outputTitles)
}

func (s *s3StoreTest) TestGetDrawing() {
	inputTitles := []string{
		"some title",
		"some other title",
	}
	inputContents := []string{
		"some content",
		"some other content",
	}

	for index := range inputTitles {
		contentReader := strings.NewReader(inputContents[index])
		errPut := s.store.PutDrawing(s.ctx, inputTitles[index], contentReader, "test-user")
		s.NoError(errPut)
	}

	content1, getContentErr1 := s.store.GetDrawing(s.ctx, inputTitles[1])
	s.NoError(getContentErr1)
	s.Equal(inputContents[1], content1)
}
