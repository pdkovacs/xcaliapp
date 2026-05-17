package awslambda

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/pdkovacs/xcaliapp/internal/s3store"
)

// The Lambda store is a single flat S3 bucket, but the web client speaks the
// multi-repo API of the full gin server. We surface that bucket as one
// synthetic repo so the client's load-time calls resolve. The name is what the
// client echoes back into subsequent /api/drawing/{repo}/{id} paths.
//
// RECONSIDER: should "repo" be in the URL at all?
//
// The repo concept earns its keep with *git* storage: a repo (better named
// "project") with an optional path lets a drawing live next to the thing it
// documents, so the storage layout mirrors the user's mental model of their
// projects. That coupling is real and worth keeping for the git backend.
//
// On S3 the setting is entirely different: the cloud abstracts the storage
// away, so "repo" carries no inherent meaning here. A hierarchy over S3
// drawings may still be a valid requirement, but its shape hasn't emerged
// yet — and baking a git-storage notion into the URL contract pre-commits an
// API we don't understand.
//
// Intended direction (separate follow-up task): a drawing's relation to other
// drawings is environment-specific metadata that belongs in the request/
// response *body*, not the URL. The path-based {repo} segment below is a
// compatibility shim for today's client, not an endorsed long-term contract;
// revisit before treating it as stable.
const (
	defaultRepoName  = "s3"
	defaultRepoLabel = "S3"
)

type drawingRepoRef struct {
	Name  string `json:"name"`
	Label string `json:"label"`
}

type drawingRepoItem struct {
	Id    string `json:"id"`
	Title string `json:"title"`
}

type drawingRepoContent struct {
	RepoRef drawingRepoRef    `json:"repoRef"`
	Items   []drawingRepoItem `json:"items"`
}

func defaultRepoRef() drawingRepoRef {
	return drawingRepoRef{Name: defaultRepoName, Label: defaultRepoLabel}
}

type putDrawingRequest struct {
	Content string `json:"content"`
}

func jsonResponse(body any) lambdaResponse {
	return lambdaResponse{
		headers: map[string]string{"Content-Type": "application/json"},
		body:    body,
	}
}

// jsonStringResponse returns body as a JSON string literal (e.g. "abc"),
// matching gin's c.JSON(200, someString). createApiGwResponse passes Go
// strings through verbatim, so we marshal here rather than handing it the raw
// value.
func jsonStringResponse(s string) (lambdaResponse, error) {
	encoded, err := json.Marshal(s)
	if err != nil {
		return lambdaResponse{}, fmt.Errorf("failed to encode string response: %w", err)
	}
	return lambdaResponse{
		headers: map[string]string{"Content-Type": "application/json"},
		body:    string(encoded),
	}, nil
}

// parseDrawingPath splits /api/drawing/{repo}[/{id}]. The synthetic single
// repo means repo is accepted but not resolved; only id matters for the store.
func parseDrawingPath(path string) (repo string, id string, ok bool) {
	rest := strings.TrimPrefix(path, "/api/drawing/")
	if rest == "" || rest == path {
		return "", "", false
	}
	parts := strings.SplitN(rest, "/", 2)
	if parts[0] == "" {
		return "", "", false
	}
	if len(parts) == 1 {
		return parts[0], "", true
	}
	return parts[0], parts[1], true
}

// extractDrawingContent reads the {"content": "..."} request body the web
// client sends for create/update. application/json is a text content type, so
// the Function-URL event carries it as a plain (non-base64) string.
func extractDrawingContent(parsedEvent map[string]any) (string, error) {
	raw, isString := parsedEvent["body"].(string)
	if !isString {
		return "", fmt.Errorf("request body is not a string: %#v", parsedEvent["body"])
	}
	var req putDrawingRequest
	if err := json.Unmarshal([]byte(raw), &req); err != nil {
		return "", fmt.Errorf("failed to unmarshal request body: %w", err)
	}
	return req.Content, nil
}

func handleListRepositoriesRequest() (lambdaResponse, error) {
	return jsonResponse([]drawingRepoRef{defaultRepoRef()}), nil
}

func handleDrawingListsRequest(ctx context.Context) (lambdaResponse, error) {
	list, listErr := drawingStore.ListDrawings(ctx)
	if listErr != nil {
		return lambdaResponse{}, fmt.Errorf("failed to list drawing titles: %w", listErr)
	}

	content := drawingRepoContent{RepoRef: defaultRepoRef(), Items: []drawingRepoItem{}}
	for id, title := range list {
		content.Items = append(content.Items, drawingRepoItem{Id: id, Title: title})
	}

	return jsonResponse(map[string]drawingRepoContent{defaultRepoName: content}), nil
}

// handleGetDrawingByPathRequest serves GET /api/drawing/{repo}/{id}. To match
// the gin server (and the client's transformResponse: JSON.parse(raw)), the
// stored content string is JSON-encoded again so the body is a JSON string
// literal the client unwraps then parses.
func handleGetDrawingByPathRequest(ctx context.Context, path string) (lambdaResponse, error) {
	_, drawingId, ok := parseDrawingPath(path)
	if !ok || drawingId == "" {
		return lambdaResponse{statusCode: http.StatusBadRequest, body: "bad drawing path"}, nil
	}

	content, getContentErr := drawingStore.GetDrawing(ctx, drawingId)
	if errors.Is(getContentErr, s3store.ErrNotfound) {
		return lambdaResponse{statusCode: http.StatusNotFound}, nil
	}
	if getContentErr != nil {
		return lambdaResponse{}, fmt.Errorf("failed to get drawing content for %s: %w", drawingId, getContentErr)
	}

	encoded, marshalErr := json.Marshal(content)
	if marshalErr != nil {
		return lambdaResponse{}, fmt.Errorf("failed to encode drawing content for %s: %w", drawingId, marshalErr)
	}

	return lambdaResponse{
		headers: map[string]string{"Content-Type": "application/json"},
		body:    string(encoded),
	}, nil
}

// handleCreateDrawingByPathRequest serves POST /api/drawing/{repo}. It mints a
// new id (matching gin's rand.Text()), stores the content, and returns the id
// as a JSON string, as the client's createDrawing mutation expects.
func handleCreateDrawingByPathRequest(ctx context.Context, path string, parsedEvent map[string]any) (lambdaResponse, error) {
	_, _, ok := parseDrawingPath(path)
	if !ok {
		return lambdaResponse{statusCode: http.StatusBadRequest, body: "bad drawing path"}, nil
	}

	content, contentErr := extractDrawingContent(parsedEvent)
	if contentErr != nil {
		return lambdaResponse{}, contentErr
	}

	drawingId := rand.Text()
	if err := drawingStore.PutDrawing(ctx, drawingId, strings.NewReader(content), emailFromContext(ctx)); err != nil {
		return lambdaResponse{}, fmt.Errorf("failed to store new drawing %s: %w", drawingId, err)
	}
	return jsonStringResponse(drawingId)
}

// handleUpdateDrawingByPathRequest serves PUT /api/drawing/{repo}/{id}, storing
// the content under the given id and echoing the id back as a JSON string.
func handleUpdateDrawingByPathRequest(ctx context.Context, path string, parsedEvent map[string]any) (lambdaResponse, error) {
	_, drawingId, ok := parseDrawingPath(path)
	if !ok || drawingId == "" {
		return lambdaResponse{statusCode: http.StatusBadRequest, body: "bad drawing path"}, nil
	}

	content, contentErr := extractDrawingContent(parsedEvent)
	if contentErr != nil {
		return lambdaResponse{}, contentErr
	}

	if err := drawingStore.PutDrawing(ctx, drawingId, strings.NewReader(content), emailFromContext(ctx)); err != nil {
		return lambdaResponse{}, fmt.Errorf("failed to store drawing %s: %w", drawingId, err)
	}
	return jsonStringResponse(drawingId)
}

// handleDeleteDrawingByPathRequest serves DELETE /api/drawing/{repo}/{id}. The
// client reads the response as text and ignores the body, so an empty 200 is
// enough (matching gin's c.Status(http.StatusOK)).
func handleDeleteDrawingByPathRequest(ctx context.Context, path string) (lambdaResponse, error) {
	_, drawingId, ok := parseDrawingPath(path)
	if !ok || drawingId == "" {
		return lambdaResponse{statusCode: http.StatusBadRequest, body: "bad drawing path"}, nil
	}

	if err := drawingStore.DeleteDrawing(ctx, drawingId, emailFromContext(ctx)); err != nil {
		return lambdaResponse{}, fmt.Errorf("failed to delete drawing %s: %w", drawingId, err)
	}
	return lambdaResponse{statusCode: http.StatusOK}, nil
}

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
		slog.WarnContext(ctx, "auth failed", "error", authErr)
		return unauthorized("Unauthorized"), nil
	}
	ctx = contextWithEmail(ctx, email)

	pathUntyped := parsedEvent["rawPath"]
	path, pathIsString := pathUntyped.(string)
	if !pathIsString {
		return LambdaResponseToAPIGW{StatusCode: http.StatusBadRequest, Body: "bad request"}, nil
	}

	method, _ := extractHTTPMethod(parsedEvent)
	slog.InfoContext(ctx, "request", "method", method, "path", path)

	var result lambdaResponse
	var handlerErr error

	switch {
	case path == "/api/drawingRepositories" && method == "GET":
		result, handlerErr = handleListRepositoriesRequest()
	case path == "/api/drawings" && method == "GET":
		result, handlerErr = handleDrawingListsRequest(ctx)
	case strings.HasPrefix(path, "/api/drawing/") && method == "GET":
		result, handlerErr = handleGetDrawingByPathRequest(ctx, path)
	case strings.HasPrefix(path, "/api/drawing/") && method == "POST":
		result, handlerErr = handleCreateDrawingByPathRequest(ctx, path, parsedEvent)
	case strings.HasPrefix(path, "/api/drawing/") && method == "PUT":
		result, handlerErr = handleUpdateDrawingByPathRequest(ctx, path, parsedEvent)
	case strings.HasPrefix(path, "/api/drawing/") && method == "DELETE":
		result, handlerErr = handleDeleteDrawingByPathRequest(ctx, path)
	case path == "/api/drawing":
		// Legacy flat query-param route; superseded by the path-based routes
		// above. Kept until the old client is fully retired.
		result, handlerErr = handleDrawingRequest(ctx, parsedEvent)
	default:
		result, handlerErr = handleServeClientRequest(ctx, event)
	}

	if handlerErr != nil {
		slog.ErrorContext(ctx, "handler error", "error", handlerErr, "method", method, "path", path)
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
