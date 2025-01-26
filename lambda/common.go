package awslambda

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"s3store"
	"strings"
)

type lambdaResponse struct {
	headers map[string]string
	body    any
}

type LambdaResponseToAPIGW struct {
	StatusCode        int                 `json:"statusCode"`
	Headers           map[string]string   `json:"headers"`
	IsBase64Encoded   bool                `json:"isBase64Encoded"`
	MultiValueHeaders map[string][]string `json:"multiValueHeaders"`
	Body              string              `json:"body"`
}

func createApiGwResponse(challange bool, session string, lambdaResult lambdaResponse) (*LambdaResponseToAPIGW, error) {
	var respStruct *LambdaResponseToAPIGW
	headers := make(map[string]string)
	bodyToSend := ""

	if challange && len(session) > 0 {
		return respStruct, fmt.Errorf("invalid arguments: either challange or session, not both")
	}

	if challange {
		return &LambdaResponseToAPIGW{
			StatusCode:        401,
			Headers:           map[string]string{"WWW-Authenticate": "Basic"},
			IsBase64Encoded:   false,
			MultiValueHeaders: nil,
			Body:              "",
		}, nil
	}

	if len(session) > 0 {
		cookieToSet := &http.Cookie{
			Name:     sessionCookieName,
			Value:    session,
			Path:     "/",
			MaxAge:   3600,
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteLaxMode,
		}

		headers = map[string]string{"Set-Cookie": cookieToSet.String()}
	}

	if lambdaResult.body != nil {
		bodyToSendInBytes, marshalErr := json.Marshal(lambdaResult.body)
		if marshalErr != nil {
			return respStruct, marshalErr
		}
		bodyToSend = string(bodyToSendInBytes)
	}

	mergedHeaders := make(map[string]string)
	for k, v := range headers {
		mergedHeaders[k] = v
	}
	for k, v := range lambdaResult.headers {
		mergedHeaders[k] = v
	}

	respStruct = &LambdaResponseToAPIGW{
		StatusCode:        200,
		Headers:           mergedHeaders,
		IsBase64Encoded:   false,
		MultiValueHeaders: nil,
		Body:              bodyToSend,
	}

	return respStruct, nil
}

func getSessionStoreBucketName() string {
	bucketName := os.Getenv("DRAWINGS_BUCKET_NAME")
	if len(bucketName) == 0 {
		panic("failed to obtain bucket-name from Lambda Context")
	}
	return bucketName
}

func initSessionStore() *s3store.SessionStore {
	store, storeErr := s3store.NewSessionStore(context.Background(), getSessionStoreBucketName())
	if storeErr != nil {
		panic(fmt.Sprintf("failed to created S3 store: %v", storeErr))
	}

	return store
}

func initDrawingStore() *s3store.DrawingStore {
	bucketName := os.Getenv("DRAWINGS_BUCKET_NAME")
	if len(bucketName) == 0 {
		panic("failed to obtain bucket-name from Lambda Context")
	}

	store, storeErr := s3store.NewDrawingStore(context.Background(), bucketName)
	if storeErr != nil {
		panic(fmt.Sprintf("failed to created S3 store: %v", storeErr))
	}

	return store
}

var (
	sessionStore = initSessionStore()
	drawingStore = initDrawingStore()
)

// parseEventCheckCreateSession parses the event and, after checking the "Cookie" and "authentication" header for credentials returns as the map-typed first parameter.
// The second return value is the session value the browser needs to set, the third parameter is an error response (most with a WWW-Authenticate challange) if any
// the last parameter is an internal processing error if any.
func parseEventCheckCreateSession(sessMan SessionManager, ctx context.Context, event json.RawMessage) (map[string]interface{}, string, *LambdaResponseToAPIGW, error) {
	var response *LambdaResponseToAPIGW
	var parsedEvent map[string]interface{}

	if eventParseErr := json.Unmarshal(event, &parsedEvent); eventParseErr != nil {
		fmt.Printf("Failed to unmarshal event: %v", eventParseErr)
		return nil, "", response, eventParseErr
	}

	fmt.Printf("parsedEvent: %#v\n", parsedEvent)
	fmt.Printf("cookies: %#v\n", parsedEvent["cookies"])
	fmt.Printf("headers: %#v\n", parsedEvent["headers"])
	fmt.Printf("multiValueHeaders: %#v\n", parsedEvent["multiValueHeaders"])
	fmt.Printf("pathParameters: %#v\n", parsedEvent["pathParameters"])

	headers, headersCastOk := parsedEvent["headers"].(map[string]any)
	if !headersCastOk {
		fmt.Printf("failed to cast headers:\n")
		return nil, "", response, fmt.Errorf("failed to cast headers")
	}

	var incomingCookieValue string
	if headers["Cookie"] != nil {
		fmt.Printf("cookies received: %#v\n", headers["Cookie"])
		cookieString, cookiesCastOk := headers["Cookie"].(string)
		if !cookiesCastOk {
			fmt.Printf("failed to cast cookies: %#v\n", headers["Cookie"])
			return nil, "", response, fmt.Errorf("failed to cast cookies")
		}
		cookies := strings.Split(cookieString, ";")
		for _, cookie := range cookies {
			cookieParts := strings.Split(cookie, "=")
			fmt.Printf("checking cookie name: %v\n", cookieParts[0])
			if cookieParts[0] == sessionCookieName {
				incomingCookieValue = cookieParts[1]
			}
		}
		if len(incomingCookieValue) == 0 {
			fmt.Printf("cookie named %s not found\n", sessionCookieName)
		}
	}

	sessionId, createSessErr := sessMan.checkCreateSession(ctx, incomingCookieValue, headers)
	if createSessErr != nil {
		fmt.Printf("failed to create session: %v\n", createSessErr)

		var challange *Challange
		if errors.As(createSessErr, &challange) {
			fmt.Printf("preparing challange for client...\n")
			response, createRespErr := createApiGwResponse(true, "", lambdaResponse{})
			if createRespErr != nil {
				fmt.Printf("failed to create response: %v\n", createRespErr)
				return nil, "", response, createRespErr
			}
			fmt.Printf("returning response with challange: %v...\n", response)
			return nil, "", response, nil
		}

		return nil, "", response, createSessErr
	}

	if len(sessionId) == 0 {
		return parsedEvent, "", response, nil
	}

	return parsedEvent, sessionId, nil, nil
}
