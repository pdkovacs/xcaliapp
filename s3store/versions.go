package s3store

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

const modifiedByMetadataKey = "modified-by"

type BlobVersion struct {
	VersionID      string
	ModifiedBy     string
	ModifiedAt     time.Time
	Size           int64
	IsLatest       bool
	IsDeleteMarker bool
}

func modifiedByMetadata(modifiedBy string) map[string]string {
	if modifiedBy == "" {
		return nil
	}
	return map[string]string{modifiedByMetadataKey: modifiedBy}
}

func (store *DrawingStore) ListVersions(ctx context.Context, title string) ([]BlobVersion, error) {
	key := fmt.Sprintf("%s/%s", drawingContentObjectKeyPrefix, title)

	input := s3.ListObjectVersionsInput{
		Bucket: &store.bucketName,
		Prefix: &key,
	}

	versions := []BlobVersion{}
	paginator := s3.NewListObjectVersionsPaginator(store.s3Client, &input)
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list object versions for %s: %w", key, err)
		}

		for _, v := range page.Versions {
			if aws.ToString(v.Key) != key {
				continue
			}
			modifiedBy, headErr := store.versionModifiedBy(ctx, key, aws.ToString(v.VersionId))
			if headErr != nil {
				return nil, fmt.Errorf("failed to read metadata for version %s of %s: %w", aws.ToString(v.VersionId), key, headErr)
			}
			versions = append(versions, BlobVersion{
				VersionID:  aws.ToString(v.VersionId),
				ModifiedBy: modifiedBy,
				ModifiedAt: aws.ToTime(v.LastModified),
				Size:       aws.ToInt64(v.Size),
				IsLatest:   aws.ToBool(v.IsLatest),
			})
		}

		for _, m := range page.DeleteMarkers {
			if aws.ToString(m.Key) != key {
				continue
			}
			versions = append(versions, BlobVersion{
				VersionID:      aws.ToString(m.VersionId),
				ModifiedAt:     aws.ToTime(m.LastModified),
				IsLatest:       aws.ToBool(m.IsLatest),
				IsDeleteMarker: true,
			})
		}
	}

	return versions, nil
}

func (store *DrawingStore) versionModifiedBy(ctx context.Context, key, versionID string) (string, error) {
	out, err := store.s3Client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket:    &store.bucketName,
		Key:       &key,
		VersionId: &versionID,
	})
	if err != nil {
		return "", err
	}
	return out.Metadata[modifiedByMetadataKey], nil
}

func (store *DrawingStore) GetVersion(ctx context.Context, title, versionID string) (string, error) {
	key := fmt.Sprintf("%s/%s", drawingContentObjectKeyPrefix, title)

	output, err := store.s3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket:    &store.bucketName,
		Key:       &key,
		VersionId: &versionID,
	})
	if err != nil {
		return "", fmt.Errorf("failed to get version %s of %s: %w", versionID, title, err)
	}
	defer output.Body.Close()

	content, readErr := io.ReadAll(output.Body)
	if readErr != nil {
		return "", fmt.Errorf("failed to read version %s of %s: %w", versionID, title, readErr)
	}
	return string(content), nil
}

func (store *DrawingStore) RestoreVersion(ctx context.Context, title, versionID, modifiedBy string) (string, error) {
	key := fmt.Sprintf("%s/%s", drawingContentObjectKeyPrefix, title)
	copySource := fmt.Sprintf("%s/%s?versionId=%s",
		store.bucketName,
		url.PathEscape(key),
		url.QueryEscape(versionID),
	)

	out, err := store.s3Client.CopyObject(ctx, &s3.CopyObjectInput{
		Bucket:            &store.bucketName,
		Key:               &key,
		CopySource:        &copySource,
		MetadataDirective: types.MetadataDirectiveReplace,
		Metadata:          modifiedByMetadata(modifiedBy),
	})
	if err != nil {
		return "", fmt.Errorf("failed to restore version %s of %s: %w", versionID, title, err)
	}
	return aws.ToString(out.VersionId), nil
}
