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

	"github.com/johnqtcg/issue2md/internal/converter"
	gh "github.com/johnqtcg/issue2md/internal/github"
)

func TestNewWebHandlerGetIndex(t *testing.T) {
	t.Parallel()

	h := newWebHandler(webDeps{
		parser:   &fakeWebParser{},
		fetcher:  &fakeWebFetcher{},
		renderer: &fakeWebRenderer{},
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "issue2md Web") {
		t.Fatalf("body should contain page title:\n%s", rec.Body.String())
	}
}

func TestNewWebHandlerConvertFlow(t *testing.T) {
	t.Parallel()

	rawURL := "https://github.com/octo/repo/issues/1"
	ref := gh.ResourceRef{Owner: "octo", Repo: "repo", Number: 1, Type: gh.ResourceIssue, URL: rawURL}
	data := gh.IssueData{
		Meta: gh.Metadata{
			Type:      gh.ResourceIssue,
			Title:     "Issue title",
			Number:    1,
			State:     "open",
			Author:    "alice",
			CreatedAt: "2026-01-01T00:00:00Z",
			UpdatedAt: "2026-01-01T00:00:00Z",
			URL:       rawURL,
		},
		Description: "Body",
	}

	h := newWebHandler(webDeps{
		parser:   &fakeWebParser{ref: ref},
		fetcher:  &fakeWebFetcher{data: data},
		renderer: &fakeWebRenderer{content: []byte("# markdown")},
	})

	form := url.Values{}
	form.Set("url", rawURL)
	req := httptest.NewRequest(http.MethodPost, "/convert", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if body := rec.Body.String(); !strings.Contains(body, "# markdown") {
		t.Fatalf("body = %q, want markdown content", body)
	}
}

func TestNewWebHandlerConvertBadRequest(t *testing.T) {
	t.Parallel()

	h := newWebHandler(webDeps{
		parser:   &fakeWebParser{err: errors.New("bad url")},
		fetcher:  &fakeWebFetcher{},
		renderer: &fakeWebRenderer{},
	})

	form := url.Values{}
	form.Set("url", "not-a-url")
	req := httptest.NewRequest(http.MethodPost, "/convert", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
}

func TestNewWebHandlerOpenAPISpecUnavailable(t *testing.T) {
	t.Parallel()

	h := newWebHandler(webDeps{
		parser:          &fakeWebParser{},
		fetcher:         &fakeWebFetcher{},
		renderer:        &fakeWebRenderer{},
		openAPISpecPath: filepath.Join(t.TempDir(), "missing.swagger.json"),
	})

	req := httptest.NewRequest(http.MethodGet, "/openapi.json", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want 503", rec.Code)
	}
}

func TestNewWebHandlerOpenAPISpecSuccess(t *testing.T) {
	t.Parallel()

	specPath := filepath.Join(t.TempDir(), "swagger.json")
	if err := os.WriteFile(specPath, []byte(`{"openapi":"3.0.0"}`), 0o600); err != nil {
		t.Fatalf("write swagger file: %v", err)
	}

	h := newWebHandler(webDeps{
		parser:          &fakeWebParser{},
		fetcher:         &fakeWebFetcher{},
		renderer:        &fakeWebRenderer{},
		openAPISpecPath: specPath,
	})

	req := httptest.NewRequest(http.MethodGet, "/openapi.json", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if got := rec.Header().Get("Content-Type"); got != "application/json; charset=utf-8" {
		t.Fatalf("content type = %q, want application/json; charset=utf-8", got)
	}
	if body := rec.Body.String(); !strings.Contains(body, `"openapi":"3.0.0"`) {
		t.Fatalf("body = %q, want swagger json payload", body)
	}
}

func TestNewWebHandlerSwaggerPageNoExternalCDNDependency(t *testing.T) {
	t.Parallel()

	h := newWebHandler(webDeps{
		parser:   &fakeWebParser{},
		fetcher:  &fakeWebFetcher{},
		renderer: &fakeWebRenderer{},
	})

	req := httptest.NewRequest(http.MethodGet, "/swagger", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	body := rec.Body.String()
	if strings.Contains(body, "unpkg.com") {
		t.Fatalf("swagger page should not depend on external CDN:\n%s", body)
	}
	if !strings.Contains(body, "/openapi.json") {
		t.Fatalf("swagger page should include local spec endpoint:\n%s", body)
	}
}

func TestNewWebHandlerConvertStatusMapping(t *testing.T) {
	t.Parallel()

	rawURL := "https://github.com/octo/repo/issues/1"
	baseRef := gh.ResourceRef{Owner: "octo", Repo: "repo", Number: 1, Type: gh.ResourceIssue, URL: rawURL}
	baseData := gh.IssueData{
		Meta: gh.Metadata{
			Type:      gh.ResourceIssue,
			Title:     "Issue title",
			Number:    1,
			State:     "open",
			Author:    "alice",
			CreatedAt: "2026-01-01T00:00:00Z",
			UpdatedAt: "2026-01-01T00:00:00Z",
			URL:       rawURL,
		},
		Description: "Body",
	}

	tcs := []struct {
		name        string
		method      string
		body        string
		contentType string
		parserErr   error
		fetchErr    error
		renderErr   error
		wantStatus  int
	}{
		{name: "method not allowed", method: http.MethodGet, wantStatus: http.StatusMethodNotAllowed},
		{name: "invalid form", method: http.MethodPost, body: "url=%zz", contentType: "application/x-www-form-urlencoded", wantStatus: http.StatusBadRequest},
		{name: "missing url", method: http.MethodPost, body: "", contentType: "application/x-www-form-urlencoded", wantStatus: http.StatusBadRequest},
		{name: "invalid github url", method: http.MethodPost, body: "url=bad", contentType: "application/x-www-form-urlencoded", parserErr: errors.New("bad url"), wantStatus: http.StatusBadRequest},
		{
			name:        "fetch not found",
			method:      http.MethodPost,
			body:        url.Values{"url": []string{rawURL}}.Encode(),
			contentType: "application/x-www-form-urlencoded",
			fetchErr:    fmt.Errorf("fetch resource: %w", gh.ErrResourceNotFound),
			wantStatus:  http.StatusNotFound,
		},
		{
			name:        "fetch auth unauthorized",
			method:      http.MethodPost,
			body:        url.Values{"url": []string{rawURL}}.Encode(),
			contentType: "application/x-www-form-urlencoded",
			fetchErr:    errors.New("http status 401: bad credentials"),
			wantStatus:  http.StatusUnauthorized,
		},
		{
			name:        "fetch auth forbidden",
			method:      http.MethodPost,
			body:        url.Values{"url": []string{rawURL}}.Encode(),
			contentType: "application/x-www-form-urlencoded",
			fetchErr:    errors.New("forbidden"),
			wantStatus:  http.StatusForbidden,
		},
		{
			name:        "fetch rate limit",
			method:      http.MethodPost,
			body:        url.Values{"url": []string{rawURL}}.Encode(),
			contentType: "application/x-www-form-urlencoded",
			fetchErr:    errors.New("http status 403: API rate limit exceeded"),
			wantStatus:  http.StatusTooManyRequests,
		},
		{
			name:        "fetch upstream failure",
			method:      http.MethodPost,
			body:        url.Values{"url": []string{rawURL}}.Encode(),
			contentType: "application/x-www-form-urlencoded",
			fetchErr:    errors.New("http status 500: upstream timeout"),
			wantStatus:  http.StatusBadGateway,
		},
		{
			name:        "render failure",
			method:      http.MethodPost,
			body:        url.Values{"url": []string{rawURL}}.Encode(),
			contentType: "application/x-www-form-urlencoded",
			renderErr:   errors.New("render failed"),
			wantStatus:  http.StatusInternalServerError,
		},
	}

	for _, tc := range tcs {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			h := newWebHandler(webDeps{
				parser:   &fakeWebParser{ref: baseRef, err: tc.parserErr},
				fetcher:  &fakeWebFetcher{data: baseData, err: tc.fetchErr},
				renderer: &fakeWebRenderer{content: []byte("# markdown"), err: tc.renderErr},
			})

			req := httptest.NewRequest(tc.method, "/convert", strings.NewReader(tc.body))
			if tc.contentType != "" {
				req.Header.Set("Content-Type", tc.contentType)
			}
			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, req)

			if rec.Code != tc.wantStatus {
				t.Fatalf("status = %d, want %d, body=%q", rec.Code, tc.wantStatus, rec.Body.String())
			}
		})
	}
}

func TestNewWebHandlerMethodNotAllowedForDocsEndpoints(t *testing.T) {
	t.Parallel()

	h := newWebHandler(webDeps{
		parser:   &fakeWebParser{},
		fetcher:  &fakeWebFetcher{},
		renderer: &fakeWebRenderer{},
	})

	tcs := []struct {
		path string
	}{
		{path: "/openapi.json"},
		{path: "/swagger"},
	}

	for _, tc := range tcs {
		tc := tc
		t.Run(tc.path, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest(http.MethodPost, tc.path, nil)
			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, req)
			if rec.Code != http.StatusMethodNotAllowed {
				t.Fatalf("status = %d, want 405", rec.Code)
			}
		})
	}
}

func TestResolveWebAddr(t *testing.T) {
	tcs := []struct {
		name string
		env  string
		want string
	}{
		{name: "default", env: "", want: ":8080"},
		{name: "custom", env: "127.0.0.1:18080", want: "127.0.0.1:18080"},
		{name: "trim spaces", env: "  :9090  ", want: ":9090"},
	}

	for _, tc := range tcs {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("ISSUE2MD_WEB_ADDR", tc.env)
			got := resolveWebAddr()
			if got != tc.want {
				t.Fatalf("resolveWebAddr() = %q, want %q", got, tc.want)
			}
		})
	}
}

type fakeWebParser struct {
	ref gh.ResourceRef
	err error
}

func (f *fakeWebParser) Parse(rawURL string) (gh.ResourceRef, error) {
	_ = rawURL
	if f.err != nil {
		return gh.ResourceRef{}, f.err
	}
	return f.ref, nil
}

type fakeWebFetcher struct {
	data gh.IssueData
	err  error
}

func (f *fakeWebFetcher) Fetch(ctx context.Context, ref gh.ResourceRef, opts gh.FetchOptions) (gh.IssueData, error) {
	_ = ctx
	_ = ref
	_ = opts
	if f.err != nil {
		return gh.IssueData{}, f.err
	}
	return f.data, nil
}

type fakeWebRenderer struct {
	content []byte
	err     error
}

func (f *fakeWebRenderer) Render(ctx context.Context, data gh.IssueData, opts converter.RenderOptions) ([]byte, error) {
	_ = ctx
	_ = data
	_ = opts
	if f.err != nil {
		return nil, f.err
	}
	return f.content, nil
}
