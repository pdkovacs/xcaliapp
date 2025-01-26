package s3store

import (
	"context"
	"crypto/rand"
	"fmt"
	"io"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/smithy-go"
)

const (
	clientObjectKey         = "client"
	credentialsObjectKey    = "credentials"
	sessionsObjectKeyPrefix = "sessions"
	sessionStoreBucketName  = "drawing-store"
)

type SessionStore struct {
	s3Client   *s3.Client
	bucketName string
}

func (store *SessionStore) GetAllowedCredentials(ctx context.Context) (string, error) {
	input := s3.GetObjectInput{
		Bucket: &store.bucketName,
		Key:    aws.String(credentialsObjectKey),
	}
	response, getErr := store.s3Client.GetObject(ctx, &input)
	if getErr != nil {
		return "", fmt.Errorf("failed to retrieve S3 Object for credentials: %w", getErr)
	}
	content, readErr := io.ReadAll(response.Body)
	if readErr != nil {
		return "", fmt.Errorf("failed to read credentials from S3: %w", readErr)
	}
	return string(content), nil
}

func (store *SessionStore) CreateSession(ctx context.Context) (string, error) {
	deleteErr := store.deleteAllSessions(ctx)
	if deleteErr != nil {
		return "", fmt.Errorf("failed to delete existing sessions while creating a new one: %w", deleteErr)
	}

	body := strings.NewReader("empty")
	sId := SessionId()
	key := fmt.Sprintf("%s/%s", sessionsObjectKeyPrefix, sId)

	input := s3.PutObjectInput{
		Bucket: &store.bucketName,
		Key:    &key,
		Body:   body,
	}
	_, err := store.s3Client.PutObject(ctx, &input)
	if err != nil {
		return "", fmt.Errorf("failed to put object for session %s: %w", key, err)
	}

	return sId, nil
}

func (store *SessionStore) ListSessions(ctx context.Context) ([]string, error) {
	return listObjectKeys(ctx, store.s3Client, store.bucketName, sessionsObjectKeyPrefix, true)
}

func (store *SessionStore) deleteAllSessions(ctx context.Context) error {
	keys, listErr := listObjectKeys(ctx, store.s3Client, store.bucketName, sessionsObjectKeyPrefix, false)
	if listErr != nil {
		return fmt.Errorf("failed to list session object keys (omitPrefix=false): %w", listErr)
	}

	if len(keys) == 0 {
		return nil
	}

	var objectIds []types.ObjectIdentifier
	for _, key := range keys {
		objectIds = append(objectIds, types.ObjectIdentifier{Key: aws.String(key)})
	}
	_, deleteErr := store.s3Client.DeleteObjects(ctx, &s3.DeleteObjectsInput{
		Bucket: aws.String(store.bucketName),
		Delete: &types.Delete{Objects: objectIds, Quiet: aws.Bool(true)},
	})
	if deleteErr != nil {
		return fmt.Errorf("failed to delete session object keys: %w", deleteErr)
	}

	return nil
}

func (store *SessionStore) ServeClientCode(ctx context.Context, path string) (string, error) {
	key := fmt.Sprintf("%s%s", clientObjectKey, path)
	fmt.Printf("ServiceClientCode: requested content from key: %s", key)
	input := s3.GetObjectInput{
		Bucket: &store.bucketName,
		Key:    aws.String(key),
	}
	response, getObjectErr := store.s3Client.GetObject(ctx, &input)

	if awsErr, ok := getObjectErr.(smithy.APIError); ok && awsErr.ErrorCode() == "NotFound" {
		return "", ErrNotfound
	}
	if getObjectErr != nil {
		return "", fmt.Errorf("failed to retrieve S3 Object for client code content: %w", getObjectErr)
	}

	content, readErr := io.ReadAll(response.Body)
	if readErr != nil {
		return "", fmt.Errorf("failed to read client code content from S3: %w", readErr)
	}
	fmt.Printf("ServiceClientCode: serving content from key: %s", key)
	return string(content), nil
}

func SessionId() string {
	buf := make([]byte, 32)

	_, err := rand.Read(buf)
	if err != nil {
		panic(err)
	}

	return fmt.Sprintf("%x", buf)
}

func NewSessionStore(ctx context.Context, bucketName string) (*SessionStore, error) {
	sdkConfig, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		fmt.Println("Couldn't load default configuration.")
		fmt.Println(err)
		return nil, fmt.Errorf("failed to load default configuration: %w", err)
	}
	s3Client := s3.NewFromConfig(sdkConfig)

	bucketNameToUse := sessionStoreBucketName
	if len(bucketName) > 0 {
		bucketNameToUse = bucketName
	}
	return &SessionStore{
		s3Client:   s3Client,
		bucketName: bucketNameToUse,
	}, nil
}
