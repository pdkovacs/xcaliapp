package gitstore

import (
	"context"
	"fmt"
	"io"
	"vcblobstore"
	"vcblobstore/git/local"

	"github.com/rs/zerolog"
)

type LocalGitStore struct {
	repo *local.Git
}

func (store *LocalGitStore) PutDrawing(ctx context.Context, title string, contentReader io.Reader, modifiedBy string) error {
	content, readErr := io.ReadAll(contentReader)
	if readErr != nil {
		return fmt.Errorf("failed to read content for %s: %w", title, readErr)
	}
	blobInfo := vcblobstore.BlobInfo{
		Key:        title,
		Content:    content,
		ModifiedBy: modifiedBy,
	}
	return store.repo.AddBlob(ctx, blobInfo)
}

func (store *LocalGitStore) ListDrawingTitles(ctx context.Context) ([]string, error) {
	return store.repo.ListBlobKeys(ctx)
}

func (store *LocalGitStore) GetDrawing(ctx context.Context, title string) (string, error) {
	blob, err := store.repo.GetBlob(ctx, title)
	if err != nil {
		return "", fmt.Errorf("failed to read drawing %s: %w", title, err)
	}
	return string(blob), nil
}

func NewLocalGitStore(pathToRepo string, logger *zerolog.Logger) (*LocalGitStore, error) {
	config := local.Config{
		Location: pathToRepo,
	}

	repo := local.NewLocalGitRepository(&config, logger)
	if createErr := repo.CreateRepository(context.TODO()); createErr != nil {
		return nil, fmt.Errorf("failed to make sure an initialized repository exists: %w", createErr)
	}

	return &LocalGitStore{
		repo: repo,
	}, nil
}
