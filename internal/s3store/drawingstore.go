package s3store

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

const (
	drawingStoreBucketName        = "drawing-store"
	drawingContentObjectKeyPrefix = "drawing-content"
	titleMetadataKey              = "title"
)

var ErrNotfound = errors.New("not found")

type DrawingStore struct {
	s3Client   *s3.Client
	bucketName string
}

func extractTitle(content []byte) string {
	var doc map[string]any
	if err := json.Unmarshal(content, &doc); err != nil {
		return ""
	}
	title, _ := doc["title"].(string)
	return title
}

func drawingMetadata(title, modifiedBy string) map[string]string {
	md := map[string]string{}
	if title != "" {
		md[titleMetadataKey] = title
	}
	if modifiedBy != "" {
		md[modifiedByMetadataKey] = modifiedBy
	}
	if len(md) == 0 {
		return nil
	}
	return md
}

func (store *DrawingStore) drawingKey(drawingId string) string {
	return fmt.Sprintf("%s/%s", drawingContentObjectKeyPrefix, drawingId)
}

func (store *DrawingStore) titleOfKey(ctx context.Context, key string) (string, error) {
	out, err := store.s3Client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: &store.bucketName,
		Key:    &key,
	})
	if err != nil {
		return "", err
	}
	return out.Metadata[titleMetadataKey], nil
}

func (store *DrawingStore) PutDrawing(ctx context.Context, drawingId string, content io.Reader, modifiedBy string) error {
	key := store.drawingKey(drawingId)

	body, err := io.ReadAll(content)
	if err != nil {
		return fmt.Errorf("failed to read content for drawing %s: %w", drawingId, err)
	}
	title := extractTitle(body)

	input := s3.PutObjectInput{
		Bucket:   &store.bucketName,
		Key:      &key,
		Body:     bytes.NewReader(body),
		Metadata: drawingMetadata(title, modifiedBy),
	}
	if _, err := store.s3Client.PutObject(ctx, &input); err != nil {
		return fmt.Errorf("failed to put object for drawing %s: %w", key, err)
	}

	return nil
}

func (store *DrawingStore) CopyDrawing(ctx context.Context, sourceId string, targetId string, modifiedBy string) error {
	sourceKey := store.drawingKey(sourceId)
	targetKey := store.drawingKey(targetId)
	copySource := fmt.Sprintf("%s/%s", store.bucketName, sourceKey)

	sourceTitle, headErr := store.titleOfKey(ctx, sourceKey)
	if headErr != nil {
		return fmt.Errorf("failed to read source title for copy %s -> %s: %w", sourceId, targetId, headErr)
	}

	input := s3.CopyObjectInput{
		Bucket:            &store.bucketName,
		Key:               &targetKey,
		CopySource:        &copySource,
		MetadataDirective: types.MetadataDirectiveReplace,
		Metadata:          drawingMetadata(sourceTitle, modifiedBy),
	}
	if _, err := store.s3Client.CopyObject(ctx, &input); err != nil {
		return fmt.Errorf("failed to copy object for drawing %s to %s: %w", copySource, targetKey, err)
	}

	return nil
}

// modifiedBy is intentionally not recorded: S3 delete-markers do not carry user
// metadata. The actor of a deletion is observable only via CloudTrail.
func (store *DrawingStore) DeleteDrawing(ctx context.Context, drawingId string, modifiedBy string) error {
	key := store.drawingKey(drawingId)

	input := s3.DeleteObjectInput{
		Bucket: &store.bucketName,
		Key:    &key,
	}
	if _, err := store.s3Client.DeleteObject(ctx, &input); err != nil {
		return fmt.Errorf("failed to delete object for drawing %s: %w", key, err)
	}

	return nil
}

func (store *DrawingStore) ListDrawings(ctx context.Context) (map[string]string, error) {
	ids, err := listObjectKeys(ctx, store.s3Client, store.bucketName, drawingContentObjectKeyPrefix, true)
	if err != nil {
		return nil, fmt.Errorf("failed to list object keys: %w", err)
	}

	drawingList := map[string]string{}
	for _, drawingId := range ids {
		title, headErr := store.titleOfKey(ctx, store.drawingKey(drawingId))
		if headErr != nil {
			return nil, fmt.Errorf("failed to read title metadata for %s: %w", drawingId, headErr)
		}
		if title == "" {
			title = drawingId
		}
		drawingList[drawingId] = title
	}

	return drawingList, nil
}

func (store *DrawingStore) GetDrawing(ctx context.Context, drawingId string) (string, error) {
	key := store.drawingKey(drawingId)
	input := s3.GetObjectInput{
		Bucket: &store.bucketName,
		Key:    &key,
	}
	output, getObjectErr := store.s3Client.GetObject(ctx, &input)
	if getObjectErr != nil {
		return "", fmt.Errorf("failed to get content object with id '%s': %w", drawingId, getObjectErr)
	}
	content, readBodyErr := io.ReadAll(output.Body)
	if readBodyErr != nil {
		return "", fmt.Errorf("failed to read content body for drawing %s: %w", drawingId, readBodyErr)
	}

	return string(content), nil
}

func listObjectKeys(ctx context.Context, s3Client *s3.Client, bucketName string, prefix string, omitPrefixFromOutput bool) ([]string, error) {
	var err error
	var output *s3.ListObjectsV2Output
	keys := []string{}
	input := s3.ListObjectsV2Input{
		Bucket: &bucketName,
		Prefix: &prefix,
	}
	objectPaginator := s3.NewListObjectsV2Paginator(s3Client, &input)
	for objectPaginator.HasMorePages() {
		output, err = objectPaginator.NextPage(ctx)
		if err != nil {
			return nil, err
		} else {
			for _, object := range output.Contents {
				keyToOutput := *object.Key
				if omitPrefixFromOutput {
					keyToOutput = string([]rune(*object.Key)[len(prefix)+1:])
				}
				keys = append(keys, keyToOutput)
			}
		}
	}
	return keys, err
}

func NewDrawingStore(ctx context.Context, bucketName string) (*DrawingStore, error) {
	sdkConfig, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		fmt.Println("Couldn't load default configuration.")
		fmt.Println(err)
		return nil, fmt.Errorf("failed to load default configuration: %w", err)
	}
	s3Client := newS3Client(sdkConfig)

	bucketNameToUse := drawingStoreBucketName
	if len(bucketName) > 0 {
		bucketNameToUse = bucketName
	}
	return &DrawingStore{
		s3Client:   s3Client,
		bucketName: bucketNameToUse,
	}, nil
}
