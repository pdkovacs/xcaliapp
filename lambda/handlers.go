package awslambda

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"s3store"
	"strings"
)

type eventHandlerFn func(parsedEvent map[string]any) (lambdaResponse, error)

func HandleSeveClientRequest(ctx context.Context, event json.RawMessage) (LambdaResponseToAPIGW, error) {
	var response LambdaResponseToAPIGW
	var parsedEvent map[string]interface{}

	if eventParseErr := json.Unmarshal(event, &parsedEvent); eventParseErr != nil {
		fmt.Printf("Failed to unmarshal event: %v", eventParseErr)
		return response, eventParseErr
	}

	fmt.Printf("parsedEvent: %#v\n", parsedEvent)

	pathAsAny := parsedEvent["path"]
	path, pathIsString := pathAsAny.(string)
	if !pathIsString {
		msg := fmt.Sprintf("event property 'path' %#v is not string", pathAsAny)
		fmt.Print(msg)
		return response, fmt.Errorf("%s", msg)
	}
	if path == "/" {
		path = "/index.html"
	}
	content, readErr := sessionStore.ServeClientCode(ctx, path)
	if readErr == s3store.ErrNotfound {
		return LambdaResponseToAPIGW{
			StatusCode: 404,
		}, nil
	}
	if readErr != nil {
		fmt.Printf("failed to read client code resource: %#v\n", readErr)
		return response, readErr
	}

	pathParts := strings.Split(path, ".")
	fmt.Printf("pathParts: %#v\n", pathParts)

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

	fmt.Printf("serving client code content as %s\n", contentType)

	return LambdaResponseToAPIGW{
		StatusCode: 200,
		Headers: map[string]string{
			"Content-Type": contentType,
		},
		Body: content,
	}, nil
}

func HandleListDrawingsRequest(ctx context.Context, event json.RawMessage) (LambdaResponseToAPIGW, error) {
	return handle(ctx, event, func(parsedEvent map[string]any) (lambdaResponse, error) {
		titles, listErr := drawingStore.ListDrawingTitles(ctx)
		if listErr != nil {
			fmt.Printf("failed to list drawing titles: %#v\n", listErr)
			return lambdaResponse{}, listErr
		}
		return lambdaResponse{
			headers: nil,
			body:    titles,
		}, nil
	})
}

func HandleGetDrawingRequest(ctx context.Context, event json.RawMessage) (LambdaResponseToAPIGW, error) {
	return handle(ctx, event, func(parsedEvent map[string]any) (lambdaResponse, error) {
		title, titleErr := extractTitlePathParam(parsedEvent)
		if titleErr != nil {
			return lambdaResponse{}, titleErr
		}

		content, getContentErr := drawingStore.GetDrawing(ctx, title)
		if getContentErr != nil {
			fmt.Printf("failed to get drawing content for %s: %#v", title, getContentErr)
			return lambdaResponse{}, getContentErr
		}

		var contentJson any
		if jsonErr := json.Unmarshal([]byte(content), &contentJson); jsonErr != nil {
			fmt.Printf("failed to unmarshal drawing content for %s into JSON: %#v", title, jsonErr)
			return lambdaResponse{}, jsonErr
		}

		fmt.Printf("content of length %d found for %s", len(content), title)
		return lambdaResponse{
			headers: map[string]string{
				"Content-Type": "application/json",
			},
			body: contentJson,
		}, nil
	})
}

func HandlePutDrawingRequest(ctx context.Context, event json.RawMessage) (LambdaResponseToAPIGW, error) {
	return handle(ctx, event, func(parsedEvent map[string]any) (lambdaResponse, error) {
		title, titleErr := extractTitlePathParam(parsedEvent)
		if titleErr != nil {
			return lambdaResponse{}, titleErr
		}
		body := parsedEvent["body"]
		content, bodyIsString := body.(string)
		if !bodyIsString {
			msg := "body for %s isn't string: %#v"
			fmt.Printf(msg+"\n", title, body)
			return lambdaResponse{}, fmt.Errorf(msg, title, body)
		}
		fmt.Printf("received content for %s of length  %d: ", title, len(content))
		contentReader := strings.NewReader(content)
		putDrawingErr := drawingStore.PutDrawing(ctx, title, contentReader, "") // TODO: modifiedBy
		if putDrawingErr != nil {
			fmt.Printf("failed to store drawing %s: %v", title, putDrawingErr)
			return lambdaResponse{}, fmt.Errorf("failed to store drawing %s: %v", title, putDrawingErr)
		}
		return lambdaResponse{}, nil
	})
}

func handle(ctx context.Context, event json.RawMessage, eventHandler eventHandlerFn) (LambdaResponseToAPIGW, error) {
	var response LambdaResponseToAPIGW

	bucketName := os.Getenv("DRAWINGS_BUCKET_NAME")
	if len(bucketName) == 0 {
		fmt.Printf("failed to obtain bucket-name from Lambda Context")
		return response, fmt.Errorf("failed to obtain bucket-name from Lambda Context")
	}

	sessMan := SessionManager{sessionStore}

	parsedEvent, sessionId, errorResponse, parseCheckErr := parseEventCheckCreateSession(sessMan, ctx, event)

	if parseCheckErr != nil {
		fmt.Printf("responding with internal error: %#v", parseCheckErr)
		return response, parseCheckErr
	}

	if errorResponse != nil {
		fmt.Printf("responding with authn error: %#v", errorResponse)
		return *errorResponse, nil
	}

	body, eventHandlerErr := eventHandler(parsedEvent)
	if eventHandlerErr != nil {
		fmt.Printf("responding with internal error: %#v", eventHandlerErr)
	}

	payloadResponse, createRespErr := createApiGwResponse(false, sessionId, body)
	if createRespErr != nil {
		fmt.Printf("failed to create response: %v\n", createRespErr)
		return response, createRespErr
	}

	fmt.Printf("handle responding with %#v: \n", *payloadResponse)

	return *payloadResponse, nil
}

func extractTitlePathParam(parsedEvent map[string]any) (string, error) {
	rawPathParameters := parsedEvent["pathParameters"]

	typedPathParams, rawPathParametersTypeOK := rawPathParameters.(map[string]any)
	if !rawPathParametersTypeOK {
		errMsg := "'pathParameters' event property is not of type map[string]string"
		fmt.Print(errMsg)
		return "", fmt.Errorf("%s", errMsg)
	}
	title, titleIsString := typedPathParams["title"].(string)
	if !titleIsString {
		errMsg := "'title' path-parameter is not string"
		fmt.Print(errMsg)
		return "", fmt.Errorf("%s", errMsg)
	}

	return title, nil
}
