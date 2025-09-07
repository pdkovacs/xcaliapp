package gitstore

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"path"
	"vcblobstore"
	"vcblobstore/git/local"

	"github.com/rs/zerolog"
)

type LocalGitRepo struct {
	blobStore      *local.Git
	pathToDrawings string
}

func (repo *LocalGitRepo) drawingIdToRepoPath(drawingId string) string {
	return path.Join(repo.pathToDrawings, drawingId)
}

func (repo *LocalGitRepo) PutDrawing(ctx context.Context, drawingId string, contentReader io.Reader, modifiedBy string) error {
	content, readErr := io.ReadAll(contentReader)
	if readErr != nil {
		return fmt.Errorf("failed to read content for %s: %w", drawingId, readErr)
	}
	blobInfo := vcblobstore.BlobInfo{
		Key:        repo.drawingIdToRepoPath(drawingId),
		Content:    content,
		ModifiedBy: modifiedBy,
	}
	return repo.blobStore.AddBlob(ctx, blobInfo)
}

func (repo *LocalGitRepo) CopyDrawing(ctx context.Context, sourceDrawingId string, destinationDrawingId string, modifiedBy string) error {
	return repo.blobStore.CopyBlob(ctx, repo.drawingIdToRepoPath(sourceDrawingId), repo.drawingIdToRepoPath(destinationDrawingId), modifiedBy)
}

func (repo *LocalGitRepo) DeleteDrawing(ctx context.Context, drawingId string, modifiedBy string) error {
	return repo.blobStore.DeleteBlob(ctx, repo.drawingIdToRepoPath(drawingId), modifiedBy)
}

func (repo *LocalGitRepo) ListDrawings(ctx context.Context) (map[string]string, error) {
	keys, err := repo.blobStore.ListBlobKeys(ctx)
	if err != nil {
		return nil, err
	}

	drawingList := map[string]string{}
	for _, key := range keys {
		if len(key) < len(repo.pathToDrawings)+1 { //|| key[0:len(repo.pathToDrawings)] != repo.pathToDrawings {
			continue
		}

		if repo.pathToDrawings != "/" && key[0:len(repo.pathToDrawings)] != repo.pathToDrawings {
			continue
		}

		fnameStart := len(repo.pathToDrawings) + 1
		if repo.pathToDrawings == "/" {
			fnameStart = 0
		}
		peeledBackKey := key[fnameStart:]
		content, err := repo.GetDrawing(ctx, string(peeledBackKey))
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

		drawingList[key[fnameStart:]] = title
	}

	return drawingList, nil
}

func (repo *LocalGitRepo) GetDrawing(ctx context.Context, drawingId string) (string, error) {
	blob, err := repo.blobStore.GetBlob(ctx, repo.drawingIdToRepoPath(drawingId))
	if err != nil {
		return "", fmt.Errorf("failed to read drawing %s: %w", drawingId, err)
	}
	return string(blob), nil
}

func NewLocalGitStore(pathToStore string, pathToDrawings string, logger *zerolog.Logger) (*LocalGitRepo, error) {
	config := local.Config{
		Location: pathToStore,
	}

	repo := local.NewLocalGitRepository(&config, logger)
	if createErr := repo.CreateRepository(context.TODO()); createErr != nil {
		return nil, fmt.Errorf("failed to make sure an initialized repository exists: %w", createErr)
	}

	return &LocalGitRepo{
		blobStore:      repo,
		pathToDrawings: pathToDrawings,
	}, nil
}
