// Command lambda-local runs the production Lambda handler behind a plain HTTP
// server for local web -> Lambda -> S3 testing (against MinIO).
//
// Every HTTP request is translated into the API Gateway / Lambda Function URL
// v2.0 event shape that awslambda.HandleRequest expects (including faithful
// base64 request-body encoding for binary content types), and the
// LambdaResponseToAPIGW it returns is translated back into an HTTP response.
//
// It is meant to be run with XCALI_DEV_AUTH=skip (bypassing the Cloudflare
// Access gate) and the MinIO AWS_* / DRAWINGS_BUCKET_NAME environment; see
// cmd/lambda-local/taskfile.yaml.
package main

import (
	"encoding/base64"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"unicode/utf8"

	awslambda "github.com/pdkovacs/xcaliapp/internal/awslambda"
)

func main() {
	addr := os.Getenv("LOCAL_LAMBDA_ADDR")
	if addr == "" {
		addr = ":8888"
	}

	http.HandleFunc("/", handle)

	log.Printf("lambda-local listening on %s (forwarding to awslambda.HandleRequest)", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

func handle(w http.ResponseWriter, r *http.Request) {
	event, err := buildEvent(r)
	if err != nil {
		http.Error(w, "failed to build event: "+err.Error(), http.StatusBadRequest)
		return
	}

	resp, err := awslambda.HandleRequest(r.Context(), event)
	if err != nil {
		log.Printf("handler error: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	writeResponse(w, resp)
}

// buildEvent renders the incoming HTTP request as the API-GW/Lambda
// Function-URL v2.0 event JSON. It mirrors the parts of the real Function-URL
// payload that affect handler behavior:
//
//   - body is base64-encoded with isBase64Encoded=true for non-text content
//     types, exactly as a real Function URL does (so handlers that read the
//     body raw will fail locally the same way they would in production);
//   - headers are lowercased and multi-values comma-joined; the Cookie header
//     is removed and split into the separate v2.0 "cookies" array;
//   - queryStringParameters comma-joins repeated keys (AWS behavior) and
//     rawQueryString is preserved.
//
// Not yet emulated (next fidelity step, intentionally omitted): the rest of
// requestContext (sourceIp, requestId, http.path/protocol, time, domainName).
// Handlers reading those will pass locally but may diverge in production.
func buildEvent(r *http.Request) (json.RawMessage, error) {
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	body, isBase64Encoded := encodeBody(bodyBytes, r.Header.Get("Content-Type"))

	headers := map[string]any{}
	for k, v := range r.Header {
		if strings.EqualFold(k, "Cookie") {
			continue // v2.0 carries cookies in a dedicated array, not headers
		}
		headers[strings.ToLower(k)] = strings.Join(v, ",")
	}

	var cookies []string
	for _, c := range r.Header.Values("Cookie") {
		for _, pair := range strings.Split(c, ";") {
			if pair = strings.TrimSpace(pair); pair != "" {
				cookies = append(cookies, pair)
			}
		}
	}

	var queryStringParameters map[string]any
	if q := r.URL.Query(); len(q) > 0 {
		queryStringParameters = map[string]any{}
		for k, vals := range q {
			queryStringParameters[k] = strings.Join(vals, ",") // AWS comma-joins repeats
		}
	}

	event := map[string]any{
		"version":               "2.0",
		"routeKey":              "$default",
		"rawPath":               r.URL.Path,
		"rawQueryString":        r.URL.RawQuery,
		"cookies":               cookies,
		"headers":               headers,
		"queryStringParameters": queryStringParameters,
		"requestContext": map[string]any{
			"http": map[string]any{"method": r.Method},
		},
		"body":            body,
		"isBase64Encoded": isBase64Encoded,
	}

	return json.Marshal(event)
}

// encodeBody reproduces the Lambda Function URL rule for the request body:
// text content types are passed through as-is, everything else (and anything
// that isn't valid UTF-8) is base64-encoded with isBase64Encoded=true.
func encodeBody(body []byte, contentType string) (string, bool) {
	if len(body) == 0 {
		return "", false
	}
	if isTextContentType(contentType) && utf8.Valid(body) {
		return string(body), false
	}
	return base64.StdEncoding.EncodeToString(body), true
}

// isTextContentType reports whether a Content-Type is treated as text (passed
// through verbatim) by a Lambda Function URL. An empty/absent Content-Type is
// treated as text, matching AWS.
func isTextContentType(contentType string) bool {
	ct := strings.ToLower(strings.TrimSpace(contentType))
	if i := strings.IndexByte(ct, ';'); i >= 0 {
		ct = strings.TrimSpace(ct[:i]) // drop ";charset=..."
	}
	switch {
	case ct == "":
		return true
	case strings.HasPrefix(ct, "text/"):
		return true
	case ct == "application/json" || strings.HasSuffix(ct, "+json"):
		return true
	case ct == "application/xml" || strings.HasSuffix(ct, "+xml"):
		return true
	case ct == "application/javascript":
		return true
	case ct == "application/x-www-form-urlencoded":
		return true
	default:
		return false
	}
}

func writeResponse(w http.ResponseWriter, resp awslambda.LambdaResponseToAPIGW) {
	for k, v := range resp.Headers {
		w.Header().Set(k, v)
	}

	status := resp.StatusCode
	if status == 0 {
		status = http.StatusOK
	}
	w.WriteHeader(status)

	if resp.Body == "" {
		return
	}
	if resp.IsBase64Encoded {
		decoded, err := base64.StdEncoding.DecodeString(resp.Body)
		if err != nil {
			log.Printf("failed to decode base64 body: %v", err)
			return
		}
		_, _ = w.Write(decoded)
		return
	}
	_, _ = io.WriteString(w, resp.Body)
}
