package s3store

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

const (
	drawingStoreBucketName        = "drawing-store"
	drawingContentObjectKeyPrefix = "drawing-content"
)

var ErrNotfound = errors.New("not found")

type DrawingStore struct {
	s3Client   *s3.Client
	bucketName string
}

// TODO: use the modifiedBy parameter
func (store *DrawingStore) PutDrawing(ctx context.Context, title string, content io.Reader, modifiedBy string) error {
	key := fmt.Sprintf("%s/%s", drawingContentObjectKeyPrefix, title)

	input := s3.PutObjectInput{
		Bucket: &store.bucketName,
		Key:    &key,
		Body:   content,
	}
	_, err := store.s3Client.PutObject(ctx, &input)
	if err != nil {
		return fmt.Errorf("failed to put object for drawing %s: %w", key, err)
	}

	return nil
}

func (store *DrawingStore) ListDrawingTitles(ctx context.Context) ([]string, error) {
	return listObjectKeys(ctx, store.s3Client, store.bucketName, drawingContentObjectKeyPrefix, true)
}

func (store *DrawingStore) GetDrawing(ctx context.Context, title string) (string, error) {
	key := fmt.Sprintf("%s/%s", drawingContentObjectKeyPrefix, title)
	input := s3.GetObjectInput{
		Bucket: &store.bucketName,
		Key:    &key,
	}
	output, getObjectErr := store.s3Client.GetObject(ctx, &input)
	if getObjectErr != nil {
		return "", fmt.Errorf("failed to get content object with title '%s': %w", title, getObjectErr)
	}
	content, readBodyErr := io.ReadAll(output.Body)
	if readBodyErr != nil {
		return "", fmt.Errorf("failed to read content body for drawing %s: %w", title, readBodyErr)
	}

	fmt.Printf(">>>>>>>>> content: %#v\n", content)

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
	s3Client := s3.NewFromConfig(sdkConfig)

	bucketNameToUse := drawingStoreBucketName
	if len(bucketName) > 0 {
		bucketNameToUse = bucketName
	}
	return &DrawingStore{
		s3Client:   s3Client,
		bucketName: bucketNameToUse,
	}, nil
}
