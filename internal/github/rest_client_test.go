package github

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"testing"
)

func TestRESTClientAuthHeader(t *testing.T) {
	t.Parallel()

	clientHTTP := newTestHTTPClient(func(r *http.Request) (*http.Response, error) {
		if r.URL.Path != "/repos/octo/repo/issues/1" {
			return notFoundResponse(r.URL.Path), nil
		}
		if got := r.Header.Get("Authorization"); got != "Bearer token-123" {
			t.Fatalf("Authorization header = %q, want %q", got, "Bearer token-123")
		}

		return mustJSONResponse(t, http.StatusOK, map[string]any{
			"number": 1,
			"title":  "hello",
		}), nil
	})

	client, err := newRESTClient(Config{
		Token:       "token-123",
		HTTPClient:  clientHTTP,
		RESTBaseURL: "https://api.test/",
	})
	if err != nil {
		t.Fatalf("newRESTClient error = %v, want nil", err)
	}

	issue, err := client.getIssue(context.Background(), "octo", "repo", 1)
	if err != nil {
		t.Fatalf("getIssue error = %v, want nil", err)
	}
	if issue.GetTitle() != "hello" {
		t.Fatalf("issue title = %q, want hello", issue.GetTitle())
	}
}

func TestRESTClientWrapsStatusError(t *testing.T) {
	t.Parallel()

	clientHTTP := newTestHTTPClient(func(r *http.Request) (*http.Response, error) {
		if r.URL.Path != "/repos/octo/repo/issues/1" {
			return notFoundResponse(r.URL.Path), nil
		}
		body, err := json.Marshal(map[string]any{
			"message": "rate limited",
		})
		if err != nil {
			t.Fatalf("marshal response body: %v", err)
		}
		return &http.Response{
			StatusCode: http.StatusTooManyRequests,
			Header: http.Header{
				"Content-Type": []string{"application/json"},
			},
			Body: io.NopCloser(bytes.NewReader(body)),
		}, nil
	})

	client, err := newRESTClient(Config{
		HTTPClient:  clientHTTP,
		RESTBaseURL: "https://api.test/",
	})
	if err != nil {
		t.Fatalf("newRESTClient error = %v, want nil", err)
	}

	_, err = client.getIssue(context.Background(), "octo", "repo", 1)
	if err == nil {
		t.Fatal("getIssue error = nil, want error")
	}

	var stErr *statusError
	if !errors.As(err, &stErr) {
		t.Fatalf("error type = %T, want *statusError", err)
	}
	if stErr.StatusCode != http.StatusTooManyRequests {
		t.Fatalf("status code = %d, want %d", stErr.StatusCode, http.StatusTooManyRequests)
	}
}
