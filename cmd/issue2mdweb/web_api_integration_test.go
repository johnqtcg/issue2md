package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/johnqtcg/issue2md/internal/converter"
	gh "github.com/johnqtcg/issue2md/internal/github"
	"github.com/johnqtcg/issue2md/internal/parser"
)

const apiIntegrationGateEnv = "ISSUE2MD_API_INTEGRATION"

func TestWebAPIIntegrationContract(t *testing.T) {
	if strings.TrimSpace(os.Getenv(apiIntegrationGateEnv)) != "1" {
		t.Skip("set ISSUE2MD_API_INTEGRATION=1 to run API integration tests")
	}

	specPath := filepath.Join(t.TempDir(), "swagger.json")
	if err := os.WriteFile(specPath, []byte(`{"swagger":"2.0","paths":{}}`), 0o600); err != nil {
		t.Fatalf("write openapi fixture: %v", err)
	}

	tmpl, err := loadTemplate()
	if err != nil {
		t.Fatalf("load template: %v", err)
	}

	handler := newWebHandler(webDeps{
		parser:          parser.New(),
		fetcher:         integrationWebFetcher{},
		renderer:        converter.NewRenderer(nil),
		tmpl:            tmpl,
		openAPISpecPath: specPath,
	})

	tcs := []struct {
		name       string
		method     string
		path       string
		body       string
		headers    map[string]string
		wantStatus int
		wantInBody string
	}{
		{
			name:       "index ok",
			method:     http.MethodGet,
			path:       "/",
			wantStatus: http.StatusOK,
			wantInBody: "issue2md Web",
		},
		{
			name:       "swagger page redirect",
			method:     http.MethodGet,
			path:       "/swagger",
			wantStatus: http.StatusTemporaryRedirect,
			wantInBody: "/swagger/index.html",
		},
		{
			name:       "swagger index ok",
			method:     http.MethodGet,
			path:       "/swagger/index.html",
			wantStatus: http.StatusOK,
			wantInBody: "Swagger UI",
		},
		{
			name:       "openapi json ok",
			method:     http.MethodGet,
			path:       "/openapi.json",
			wantStatus: http.StatusOK,
			wantInBody: `"swagger":"2.0"`,
		},
		{
			name:       "convert missing url",
			method:     http.MethodPost,
			path:       "/convert",
			body:       "",
			headers:    map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
			wantStatus: http.StatusBadRequest,
			wantInBody: "missing url",
		},
		{
			name:       "convert invalid url",
			method:     http.MethodPost,
			path:       "/convert",
			body:       "url=bad",
			headers:    map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
			wantStatus: http.StatusBadRequest,
			wantInBody: "invalid github url",
		},
		{
			name:       "convert success",
			method:     http.MethodPost,
			path:       "/convert",
			body:       url.Values{"url": []string{"https://github.com/octo/repo/issues/1"}}.Encode(),
			headers:    map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
			wantStatus: http.StatusOK,
			wantInBody: "## Metadata",
		},
		{
			name:       "convert not found",
			method:     http.MethodPost,
			path:       "/convert",
			body:       url.Values{"url": []string{"https://github.com/octo/repo/issues/404"}}.Encode(),
			headers:    map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
			wantStatus: http.StatusNotFound,
			wantInBody: "fetch github resource failed",
		},
		{
			name:       "convert unauthorized",
			method:     http.MethodPost,
			path:       "/convert",
			body:       url.Values{"url": []string{"https://github.com/octo/repo/issues/401"}}.Encode(),
			headers:    map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
			wantStatus: http.StatusUnauthorized,
			wantInBody: "fetch github resource failed",
		},
		{
			name:       "convert forbidden",
			method:     http.MethodPost,
			path:       "/convert",
			body:       url.Values{"url": []string{"https://github.com/octo/repo/issues/403"}}.Encode(),
			headers:    map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
			wantStatus: http.StatusForbidden,
			wantInBody: "fetch github resource failed",
		},
		{
			name:       "convert rate limited",
			method:     http.MethodPost,
			path:       "/convert",
			body:       url.Values{"url": []string{"https://github.com/octo/repo/issues/429"}}.Encode(),
			headers:    map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
			wantStatus: http.StatusTooManyRequests,
			wantInBody: "fetch github resource failed",
		},
		{
			name:       "convert upstream error",
			method:     http.MethodPost,
			path:       "/convert",
			body:       url.Values{"url": []string{"https://github.com/octo/repo/issues/500"}}.Encode(),
			headers:    map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
			wantStatus: http.StatusBadGateway,
			wantInBody: "fetch github resource failed",
		},
		{
			name:       "docs method not allowed",
			method:     http.MethodPost,
			path:       "/swagger",
			wantStatus: http.StatusMethodNotAllowed,
			wantInBody: "method not allowed",
		},
	}

	for _, tc := range tcs {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			reqCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()

			req := httptest.NewRequest(tc.method, tc.path, strings.NewReader(tc.body))
			req = req.WithContext(reqCtx)
			for k, v := range tc.headers {
				req.Header.Set(k, v)
			}

			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)
			body := rec.Body.String()

			if rec.Code != tc.wantStatus {
				t.Fatalf("status=%d want=%d body=%q", rec.Code, tc.wantStatus, body)
			}
			if tc.wantInBody != "" && !strings.Contains(body, tc.wantInBody) {
				t.Fatalf("body=%q should contain %q", body, tc.wantInBody)
			}
		})
	}
}

type integrationWebFetcher struct{}

func (integrationWebFetcher) Fetch(ctx context.Context, ref gh.ResourceRef, opts gh.FetchOptions) (gh.IssueData, error) {
	if !opts.IncludeComments {
		return gh.IssueData{}, fmt.Errorf("integration contract expects IncludeComments=true")
	}
	if err := ctx.Err(); err != nil {
		return gh.IssueData{}, err
	}

	switch ref.Number {
	case 1:
		return integrationIssueData(ref.URL), nil
	case 404:
		return gh.IssueData{}, fmt.Errorf("resource not found: %w", gh.ErrResourceNotFound)
	case 401:
		return gh.IssueData{}, errors.New("http status 401: bad credentials")
	case 403:
		return gh.IssueData{}, errors.New("forbidden")
	case 429:
		return gh.IssueData{}, errors.New("http status 403: API rate limit exceeded")
	case 500:
		return gh.IssueData{}, errors.New("http status 500: upstream timeout")
	default:
		return integrationIssueData(ref.URL), nil
	}
}

func integrationIssueData(rawURL string) gh.IssueData {
	return gh.IssueData{
		Meta: gh.Metadata{
			Type:      gh.ResourceIssue,
			Title:     "API Integration Issue",
			Number:    1,
			State:     "open",
			Author:    "integration-bot",
			CreatedAt: "2026-02-01T00:00:00Z",
			UpdatedAt: "2026-02-01T00:00:00Z",
			URL:       rawURL,
		},
		Description: "integration description",
		Thread: []gh.CommentNode{
			{
				ID:        "c1",
				Author:    "alice",
				Body:      "looks good",
				CreatedAt: "2026-02-01T00:01:00Z",
				UpdatedAt: "2026-02-01T00:01:00Z",
				URL:       rawURL + "#issuecomment-1",
			},
		},
	}
}
