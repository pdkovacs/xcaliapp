package awslambda

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
)

func HandleEcho(ctx context.Context, event json.RawMessage) (LambdaResponseToAPIGW, error) {
	var response LambdaResponseToAPIGW

	bucketName := os.Getenv("DRAWINGS_BUCKET_NAME")
	if len(bucketName) == 0 {
		fmt.Printf("failed to obtain bucket-name from Lambda Context")
		return response, fmt.Errorf("failed to obtain bucket-name from Lambda Context")
	}

	sessMan := SessionManager{sessionStore}

	_, sessionId, errorResponse, parseCheckErr := parseEventCheckCreateSession(sessMan, ctx, event)

	if parseCheckErr != nil {
		fmt.Printf("responding with itnernal error: %#v", parseCheckErr)
		return response, parseCheckErr
	}

	if errorResponse != nil {
		fmt.Printf("responding with authn error: %#v", errorResponse)
		return *errorResponse, nil
	}

	body := map[string]string{"message": "hello, xcali!"}

	payloadResponse, createRespErr := createApiGwResponse(false, sessionId, lambdaResponse{body: body})
	if createRespErr != nil {
		fmt.Printf("failed to create response: %v\n", createRespErr)
		return response, createRespErr
	}

	return *payloadResponse, nil
}
