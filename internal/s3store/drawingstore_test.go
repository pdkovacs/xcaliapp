package s3store

import (
	"context"
	"fmt"
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

func drawingJSON(title string) string {
	return fmt.Sprintf(`{"type":"excalidraw","title":%q,"elements":[]}`, title)
}

func (s *s3StoreTest) TestListDrawings() {
	drawings := []struct {
		id    string
		title string
	}{
		{"id-A", "some title"},
		{"id-B", "some other title"},
	}

	for _, d := range drawings {
		errPut := s.store.PutDrawing(s.ctx, d.id, strings.NewReader(drawingJSON(d.title)), "test-user")
		s.NoError(errPut)
	}

	drawingList, listErr := s.store.ListDrawings(s.ctx)
	s.NoError(listErr)

	expected := map[string]string{
		"id-A": "some title",
		"id-B": "some other title",
	}
	s.Equal(expected, drawingList)
}

func (s *s3StoreTest) TestGetDrawing() {
	drawings := []struct {
		id    string
		title string
	}{
		{"id-A", "some title"},
		{"id-B", "some other title"},
	}

	for _, d := range drawings {
		errPut := s.store.PutDrawing(s.ctx, d.id, strings.NewReader(drawingJSON(d.title)), "test-user")
		s.NoError(errPut)
	}

	content, getContentErr := s.store.GetDrawing(s.ctx, drawings[1].id)
	s.NoError(getContentErr)
	s.Equal(drawingJSON(drawings[1].title), content)
}
