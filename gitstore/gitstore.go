package gitstore

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"vcblobstore"
	"vcblobstore/git/local"

	"github.com/rs/zerolog"
)

type LocalGitStore struct {
	repo *local.Git
}

func (store *LocalGitStore) PutDrawing(ctx context.Context, key string, contentReader io.Reader, modifiedBy string) error {
	content, readErr := io.ReadAll(contentReader)
	if readErr != nil {
		return fmt.Errorf("failed to read content for %s: %w", key, readErr)
	}
	blobInfo := vcblobstore.BlobInfo{
		Key:        key,
		Content:    content,
		ModifiedBy: modifiedBy,
	}
	return store.repo.AddBlob(ctx, blobInfo)
}

func (store *LocalGitStore) CopyDrawing(ctx context.Context, sourcekey string, destinationkey string, modifiedBy string) error {
	return store.repo.CopyBlob(ctx, sourcekey, destinationkey, modifiedBy)
}

func (store *LocalGitStore) DeleteDrawing(ctx context.Context, title string, modifiedBy string) error {
	return store.repo.DeleteBlob(ctx, title, modifiedBy)
}

func (store *LocalGitStore) ListDrawings(ctx context.Context) (map[string]string, error) {
	keys, err := store.repo.ListBlobKeys(ctx)
	if err != nil {
		return nil, err
	}

	drawingList := map[string]string{}
	for _, key := range keys {
		content, err := store.GetDrawing(ctx, key)
		if err != nil {
			return nil, fmt.Errorf("failed to get drawing %s: %w", key, err)
		}

		var drawingContent map[string]any
		err = json.Unmarshal([]byte(content), &drawingContent)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal drawing %s: %w", key, err)
		}

		titleAny, ok := drawingContent["title"]
		if !ok {
			return nil, fmt.Errorf("drawing title not found for %s", key)
		}
		title, cast := titleAny.(string)
		if !cast {
			return nil, fmt.Errorf("title of %s is not a string: '%T'", key, titleAny)
		}
		drawingList[key] = title
	}

	return drawingList, nil
}

func (store *LocalGitStore) GetDrawing(ctx context.Context, key string) (string, error) {
	blob, err := store.repo.GetBlob(ctx, key)
	if err != nil {
		return "", fmt.Errorf("failed to read drawing %s: %w", key, err)
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
