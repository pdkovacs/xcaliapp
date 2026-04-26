package awslambda

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"s3store"
	"strings"
)

func handleServeClientRequest(ctx context.Context, event json.RawMessage) (lambdaResponse, error) {
	var parsedEvent map[string]any
	if err := json.Unmarshal(event, &parsedEvent); err != nil {
		return lambdaResponse{}, fmt.Errorf("failed to unmarshal event: %w", err)
	}

	pathAsAny := parsedEvent["rawPath"]
	path, pathIsString := pathAsAny.(string)
	if !pathIsString {
		return lambdaResponse{}, fmt.Errorf("event property 'rawPath' %#v is not string", pathAsAny)
	}
	if path == "/" {
		path = "/index.html"
	}
	content, readErr := sessionStore.ServeClientCode(ctx, path)
	if errors.Is(readErr, s3store.ErrNotfound) {
		return lambdaResponse{statusCode: http.StatusNotFound}, nil
	}
	if readErr != nil {
		return lambdaResponse{}, fmt.Errorf("failed to read client code resource: %w", readErr)
	}

	pathParts := strings.Split(path, ".")
	extension := pathParts[len(pathParts)-1]
	contentType := fmt.Sprintf("font/%s", extension)

	switch extension {
	case "html":
		contentType = "text/html"
	case "js":
		contentType = "text/javascript"
	case "css":
		contentType = "text/css"
	}

	return lambdaResponse{
		headers: map[string]string{"Content-Type": contentType},
		body:    content,
	}, nil
}

func handleListDrawingsRequest(ctx context.Context) (lambdaResponse, error) {
	titles, listErr := drawingStore.ListDrawings(ctx)
	if listErr != nil {
		return lambdaResponse{}, fmt.Errorf("failed to list drawing titles: %w", listErr)
	}
	return lambdaResponse{body: titles}, nil
}

func handleGetDrawingRequest(ctx context.Context, drawingId string) (lambdaResponse, error) {
	content, getContentErr := drawingStore.GetDrawing(ctx, drawingId)
	if getContentErr != nil {
		return lambdaResponse{}, fmt.Errorf("failed to get drawing content for %s: %w", drawingId, getContentErr)
	}

	var contentJson any
	if jsonErr := json.Unmarshal([]byte(content), &contentJson); jsonErr != nil {
		return lambdaResponse{}, fmt.Errorf("failed to unmarshal drawing content for %s: %w", drawingId, jsonErr)
	}

	return lambdaResponse{
		headers: map[string]string{"Content-Type": "application/json"},
		body:    contentJson,
	}, nil
}

func handlePutDrawingRequest(ctx context.Context, parsedEvent map[string]any) (lambdaResponse, error) {
	drawingId, idErr := extractIdQueryParam(parsedEvent)
	if idErr != nil {
		return lambdaResponse{}, idErr
	}
	body := parsedEvent["body"]
	content, bodyIsString := body.(string)
	if !bodyIsString {
		return lambdaResponse{}, fmt.Errorf("body for %s isn't string: %#v", drawingId, body)
	}
	contentReader := strings.NewReader(content)
	if err := drawingStore.PutDrawing(ctx, drawingId, contentReader, emailFromContext(ctx)); err != nil {
		return lambdaResponse{}, fmt.Errorf("failed to store drawing %s: %w", drawingId, err)
	}
	return lambdaResponse{}, nil
}

func handleDrawingRequest(ctx context.Context, parsedEvent map[string]any) (lambdaResponse, error) {
	httpMethod, httpMethodErr := extractHTTPMethod(parsedEvent)
	if httpMethodErr != nil {
		return lambdaResponse{}, httpMethodErr
	}

	if httpMethod == "GET" {
		drawingId, idErr := extractIdQueryParam(parsedEvent)
		if idErr != nil {
			return lambdaResponse{}, idErr
		}
		if len(drawingId) == 0 {
			return handleListDrawingsRequest(ctx)
		}
		return handleGetDrawingRequest(ctx, drawingId)
	}

	if httpMethod == "PUT" {
		return handlePutDrawingRequest(ctx, parsedEvent)
	}

	return lambdaResponse{}, fmt.Errorf("unexpected httpMethod: %s", httpMethod)
}

func HandleRequest(ctx context.Context, event json.RawMessage) (LambdaResponseToAPIGW, error) {
	parsedEvent, email, authErr := parseEventVerifyAccess(event)
	if authErr != nil {
		fmt.Printf("auth failed: %v\n", authErr)
		return unauthorized("Unauthorized"), nil
	}
	ctx = contextWithEmail(ctx, email)

	pathUntyped := parsedEvent["rawPath"]
	path, pathIsString := pathUntyped.(string)
	if !pathIsString {
		return LambdaResponseToAPIGW{StatusCode: http.StatusBadRequest, Body: "bad request"}, nil
	}

	var result lambdaResponse
	var handlerErr error

	if path == "/api/drawing" {
		result, handlerErr = handleDrawingRequest(ctx, parsedEvent)
	} else {
		result, handlerErr = handleServeClientRequest(ctx, event)
	}

	if handlerErr != nil {
		fmt.Printf("handler error: %v\n", handlerErr)
		return LambdaResponseToAPIGW{StatusCode: http.StatusInternalServerError, Body: "internal error"}, nil
	}

	response, respErr := createApiGwResponse(result)
	if respErr != nil {
		return LambdaResponseToAPIGW{StatusCode: http.StatusInternalServerError}, nil
	}
	return *response, nil
}

func extractHTTPMethod(parsedEvent map[string]any) (string, error) {
	requestContext, ok := parsedEvent["requestContext"].(map[string]any)
	if !ok {
		return "", fmt.Errorf("event has no requestContext")
	}
	httpInfo, ok := requestContext["http"].(map[string]any)
	if !ok {
		return "", fmt.Errorf("requestContext has no http")
	}
	method, ok := httpInfo["method"].(string)
	if !ok {
		return "", fmt.Errorf("requestContext.http.method is not string: %#v", httpInfo["method"])
	}
	return method, nil
}

func extractIdQueryParam(parsedEvent map[string]any) (string, error) {
	rawQueryParameters := parsedEvent["queryStringParameters"]
	if rawQueryParameters == nil {
		return "", nil
	}

	typedQueryParams, ok := rawQueryParameters.(map[string]any)
	if !ok {
		return "", fmt.Errorf("'queryStringParameters' event property is not of type map[string]any")
	}

	untypedIdParam := typedQueryParams["id"]
	if untypedIdParam == nil {
		return "", nil
	}

	drawingId, idIsString := untypedIdParam.(string)
	if !idIsString {
		return "", fmt.Errorf("'id' query-parameter is not string")
	}

	return drawingId, nil
}
