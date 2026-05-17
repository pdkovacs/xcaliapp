package main

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func mustBuild(t *testing.T, r *http.Request) []byte {
	t.Helper()
	raw, err := buildEvent(r)
	if err != nil {
		t.Fatalf("buildEvent: %v", err)
	}
	return raw
}

func decodeEvent(t *testing.T, raw []byte) map[string]any {
	t.Helper()
	var ev map[string]any
	if err := json.Unmarshal(raw, &ev); err != nil {
		t.Fatalf("event is not valid JSON: %v", err)
	}
	return ev
}

func TestBuildEvent_TextBodyPassthrough(t *testing.T) {
	drawing := `{"type":"excalidraw","title":"t","elements":[]}`
	r := httptest.NewRequest("PUT", "/api/drawing?id=abc", strings.NewReader(drawing))
	r.Header.Set("Content-Type", "application/json")

	raw, err := buildEvent(r)
	if err != nil {
		t.Fatalf("buildEvent: %v", err)
	}
	ev := decodeEvent(t, raw)

	if ev["isBase64Encoded"] != false {
		t.Errorf("isBase64Encoded = %v, want false for JSON", ev["isBase64Encoded"])
	}
	if ev["body"] != drawing {
		t.Errorf("body = %q, want verbatim JSON", ev["body"])
	}
	if ev["version"] != "2.0" || ev["routeKey"] != "$default" {
		t.Errorf("missing v2.0 envelope: version=%v routeKey=%v", ev["version"], ev["routeKey"])
	}
	if ev["rawQueryString"] != "id=abc" {
		t.Errorf("rawQueryString = %v, want id=abc", ev["rawQueryString"])
	}
}

func TestBuildEvent_BinaryBodyIsBase64(t *testing.T) {
	bin := []byte{0x00, 0xff, 0x10, 0x80, 0x7f}
	r := httptest.NewRequest("PUT", "/api/drawing", strings.NewReader(string(bin)))
	r.Header.Set("Content-Type", "application/octet-stream")

	raw, err := buildEvent(r)
	if err != nil {
		t.Fatalf("buildEvent: %v", err)
	}
	ev := decodeEvent(t, raw)

	if ev["isBase64Encoded"] != true {
		t.Fatalf("isBase64Encoded = %v, want true for binary", ev["isBase64Encoded"])
	}
	got, err := base64.StdEncoding.DecodeString(ev["body"].(string))
	if err != nil {
		t.Fatalf("body is not valid base64: %v", err)
	}
	if string(got) != string(bin) {
		t.Errorf("decoded body = %v, want %v", got, bin)
	}
}

func TestBuildEvent_EmptyBody(t *testing.T) {
	r := httptest.NewRequest("GET", "/api/drawing", nil)
	ev := decodeEvent(t, mustBuild(t, r))
	if ev["body"] != "" || ev["isBase64Encoded"] != false {
		t.Errorf("empty body: body=%q isBase64=%v", ev["body"], ev["isBase64Encoded"])
	}
}

func TestBuildEvent_CookiesSplitOutOfHeaders(t *testing.T) {
	r := httptest.NewRequest("GET", "/api/drawing", nil)
	r.Header.Set("Cookie", "a=1; b=2")

	ev := decodeEvent(t, mustBuild(t, r))

	headers := ev["headers"].(map[string]any)
	if _, present := headers["cookie"]; present {
		t.Error("Cookie must be stripped from headers in v2.0")
	}
	cookies, _ := ev["cookies"].([]any)
	if len(cookies) != 2 || cookies[0] != "a=1" || cookies[1] != "b=2" {
		t.Errorf("cookies = %v, want [a=1 b=2]", cookies)
	}
}

func TestBuildEvent_MultiValueQueryCommaJoined(t *testing.T) {
	r := httptest.NewRequest("GET", "/api/drawing?id=x&id=y", nil)
	ev := decodeEvent(t, mustBuild(t, r))
	qsp := ev["queryStringParameters"].(map[string]any)
	if qsp["id"] != "x,y" {
		t.Errorf("queryStringParameters[id] = %v, want x,y", qsp["id"])
	}
}

func TestIsTextContentType(t *testing.T) {
	text := []string{"", "text/html", "application/json", "application/json; charset=utf-8",
		"application/vnd.api+json", "image/svg+xml", "application/x-www-form-urlencoded"}
	binary := []string{"application/octet-stream", "image/png", "application/pdf"}
	for _, ct := range text {
		if !isTextContentType(ct) {
			t.Errorf("isTextContentType(%q) = false, want true", ct)
		}
	}
	for _, ct := range binary {
		if isTextContentType(ct) {
			t.Errorf("isTextContentType(%q) = true, want false", ct)
		}
	}
}
