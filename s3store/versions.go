package s3store

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"vcblobstore"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

const modifiedByMetadataKey = "modified-by"

func (store *DrawingStore) versionMetadata(ctx context.Context, key, versionID string) (map[string]string, error) {
	out, err := store.s3Client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket:    &store.bucketName,
		Key:       &key,
		VersionId: &versionID,
	})
	if err != nil {
		return nil, err
	}
	return out.Metadata, nil
}

func (store *DrawingStore) ListVersions(ctx context.Context, drawingId string) ([]vcblobstore.BlobVersion, error) {
	key := store.drawingKey(drawingId)

	input := s3.ListObjectVersionsInput{
		Bucket: &store.bucketName,
		Prefix: &key,
	}

	versions := []vcblobstore.BlobVersion{}
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
			md, headErr := store.versionMetadata(ctx, key, aws.ToString(v.VersionId))
			if headErr != nil {
				return nil, fmt.Errorf("failed to read metadata for version %s of %s: %w", aws.ToString(v.VersionId), key, headErr)
			}
			versions = append(versions, vcblobstore.BlobVersion{
				VersionID:  aws.ToString(v.VersionId),
				ModifiedBy: md[modifiedByMetadataKey],
				ModifiedAt: aws.ToTime(v.LastModified),
				Size:       aws.ToInt64(v.Size),
				IsLatest:   aws.ToBool(v.IsLatest),
			})
		}

		for _, m := range page.DeleteMarkers {
			if aws.ToString(m.Key) != key {
				continue
			}
			versions = append(versions, vcblobstore.BlobVersion{
				VersionID:      aws.ToString(m.VersionId),
				ModifiedAt:     aws.ToTime(m.LastModified),
				IsLatest:       aws.ToBool(m.IsLatest),
				IsDeleteMarker: true,
			})
		}
	}

	return versions, nil
}

func (store *DrawingStore) GetVersion(ctx context.Context, drawingId, versionID string) (string, error) {
	key := store.drawingKey(drawingId)

	output, err := store.s3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket:    &store.bucketName,
		Key:       &key,
		VersionId: &versionID,
	})
	if err != nil {
		return "", fmt.Errorf("failed to get version %s of %s: %w", versionID, drawingId, err)
	}
	defer output.Body.Close()

	content, readErr := io.ReadAll(output.Body)
	if readErr != nil {
		return "", fmt.Errorf("failed to read version %s of %s: %w", versionID, drawingId, readErr)
	}
	return string(content), nil
}

func (store *DrawingStore) RestoreVersion(ctx context.Context, drawingId, versionID, modifiedBy string) (string, error) {
	key := store.drawingKey(drawingId)
	copySource := fmt.Sprintf("%s/%s?versionId=%s",
		store.bucketName,
		url.PathEscape(key),
		url.QueryEscape(versionID),
	)

	sourceMd, headErr := store.versionMetadata(ctx, key, versionID)
	if headErr != nil {
		return "", fmt.Errorf("failed to read metadata for version %s of %s: %w", versionID, drawingId, headErr)
	}

	out, err := store.s3Client.CopyObject(ctx, &s3.CopyObjectInput{
		Bucket:            &store.bucketName,
		Key:               &key,
		CopySource:        &copySource,
		MetadataDirective: types.MetadataDirectiveReplace,
		Metadata:          drawingMetadata(sourceMd[titleMetadataKey], modifiedBy),
	})
	if err != nil {
		return "", fmt.Errorf("failed to restore version %s of %s: %w", versionID, drawingId, err)
	}
	return aws.ToString(out.VersionId), nil
}
