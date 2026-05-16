package awslambda

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleEcho(ctx context.Context, event json.RawMessage) (LambdaResponseToAPIGW, error) {
	_, email, authErr := parseEventVerifyAccess(event)
	if authErr != nil {
		fmt.Printf("auth failed: %v\n", authErr)
		return unauthorized("Unauthorized"), nil
	}

	body := map[string]string{
		"message": "hello, xcali!",
		"email":   email,
	}

	response, err := createApiGwResponse(lambdaResponse{body: body})
	if err != nil {
		return LambdaResponseToAPIGW{StatusCode: http.StatusInternalServerError}, nil
	}
	return *response, nil
}
