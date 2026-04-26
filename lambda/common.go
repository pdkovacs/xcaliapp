package awslambda

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"s3store"
)

type lambdaResponse struct {
	statusCode int
	headers    map[string]string
	body       any
}

type LambdaResponseToAPIGW struct {
	StatusCode        int                 `json:"statusCode"`
	Headers           map[string]string   `json:"headers"`
	IsBase64Encoded   bool                `json:"isBase64Encoded"`
	MultiValueHeaders map[string][]string `json:"multiValueHeaders"`
	Body              string              `json:"body"`
}

func unauthorized(message string) LambdaResponseToAPIGW {
	return LambdaResponseToAPIGW{
		StatusCode: http.StatusUnauthorized,
		Headers:    map[string]string{"Content-Type": "text/plain"},
		Body:       message,
	}
}

func createApiGwResponse(result lambdaResponse) (*LambdaResponseToAPIGW, error) {
	bodyToSend := ""
	if result.body != nil {
		if s, isString := result.body.(string); isString {
			bodyToSend = s
		} else {
			marshalled, marshalErr := json.Marshal(result.body)
			if marshalErr != nil {
				return nil, marshalErr
			}
			bodyToSend = string(marshalled)
		}
	}

	status := http.StatusOK
	if result.statusCode > 0 {
		status = result.statusCode
	}

	return &LambdaResponseToAPIGW{
		StatusCode:      status,
		Headers:         result.headers,
		IsBase64Encoded: false,
		Body:            bodyToSend,
	}, nil
}

func getDrawingsBucketName() string {
	bucketName := os.Getenv("DRAWINGS_BUCKET_NAME")
	if bucketName == "" {
		panic("DRAWINGS_BUCKET_NAME must be set")
	}
	return bucketName
}

func initSessionStore() *s3store.SessionStore {
	store, err := s3store.NewSessionStore(context.Background(), getDrawingsBucketName())
	if err != nil {
		panic(fmt.Sprintf("failed to create session store: %v", err))
	}
	return store
}

func initDrawingStore() *s3store.DrawingStore {
	store, err := s3store.NewDrawingStore(context.Background(), getDrawingsBucketName())
	if err != nil {
		panic(fmt.Sprintf("failed to create drawing store: %v", err))
	}
	return store
}

var (
	sessionStore = initSessionStore()
	drawingStore = initDrawingStore()
)

func parseEventVerifyAccess(event json.RawMessage) (map[string]any, string, error) {
	var parsedEvent map[string]any
	if err := json.Unmarshal(event, &parsedEvent); err != nil {
		return nil, "", fmt.Errorf("failed to unmarshal event: %w", err)
	}

	headers, ok := parsedEvent["headers"].(map[string]any)
	if !ok {
		return nil, "", fmt.Errorf("event has no headers")
	}

	if cfIP, _ := headers["cf-connecting-ip"].(string); cfIP == "" {
		return nil, "", fmt.Errorf("missing cf-connecting-ip header")
	}

	var token string
	if raw, present := headers[accessJWTHeaderKey]; present {
		if s, isString := raw.(string); isString {
			token = s
		}
	}
	if token == "" {
		return nil, "", fmt.Errorf("missing %s header", accessJWTHeaderKey)
	}

	email, verifyErr := verifier.verify(token)
	if verifyErr != nil {
		return nil, "", verifyErr
	}

	return parsedEvent, email, nil
}
