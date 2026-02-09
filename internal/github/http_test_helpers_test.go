package github

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
)

type roundTripFunc func(req *http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func newTestHTTPClient(fn roundTripFunc) *http.Client {
	return &http.Client{
		Transport: fn,
	}
}

func jsonHTTPResponse(statusCode int, payload any) (*http.Response, error) {
	buf := bytes.NewBuffer(nil)
	if err := json.NewEncoder(buf).Encode(payload); err != nil {
		return nil, err
	}
	return &http.Response{
		StatusCode: statusCode,
		Header: http.Header{
			"Content-Type": []string{"application/json"},
		},
		Body: io.NopCloser(buf),
	}, nil
}

func textHTTPResponse(statusCode int, body string) *http.Response {
	return &http.Response{
		StatusCode: statusCode,
		Header:     http.Header{},
		Body:       io.NopCloser(strings.NewReader(body)),
	}
}

func mustJSONResponse(t *testing.T, statusCode int, payload any) *http.Response {
	t.Helper()
	resp, err := jsonHTTPResponse(statusCode, payload)
	if err != nil {
		t.Fatalf("build json response: %v", err)
	}
	return resp
}

func notFoundResponse(path string) *http.Response {
	return textHTTPResponse(http.StatusNotFound, fmt.Sprintf(`{"message":"not found: %s"}`, path))
}
