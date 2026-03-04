package webapp

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

func TestNewHandlerGetIndex(t *testing.T) {
	t.Parallel()

	h := NewHandler(Deps{
		Parser:   &fakeWebParser{},
		Fetcher:  &fakeWebFetcher{},
		Renderer: &fakeWebRenderer{},
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

func TestNewHandlerConvertFlow(t *testing.T) {
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

	h := NewHandler(Deps{
		Parser:   &fakeWebParser{ref: ref},
		Fetcher:  &fakeWebFetcher{data: data},
		Renderer: &fakeWebRenderer{content: []byte("# markdown")},
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
	if got := rec.Header().Get("Content-Type"); got != "text/plain; charset=utf-8" {
		t.Fatalf("content type = %q, want text/plain; charset=utf-8", got)
	}
}

func TestNewHandlerOpenAPISpecUnavailable(t *testing.T) {
	t.Parallel()

	h := NewHandler(Deps{
		Parser:          &fakeWebParser{},
		Fetcher:         &fakeWebFetcher{},
		Renderer:        &fakeWebRenderer{},
		OpenAPISpecPath: filepath.Join(t.TempDir(), "missing.swagger.json"),
	})

	req := httptest.NewRequest(http.MethodGet, "/openapi.json", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want 503", rec.Code)
	}
}

func TestNewHandlerOpenAPISpecSuccess(t *testing.T) {
	t.Parallel()

	specPath := filepath.Join(t.TempDir(), "swagger.json")
	if err := os.WriteFile(specPath, []byte(`{"openapi":"3.0.0"}`), 0o600); err != nil {
		t.Fatalf("write swagger file: %v", err)
	}

	h := NewHandler(Deps{
		Parser:          &fakeWebParser{},
		Fetcher:         &fakeWebFetcher{},
		Renderer:        &fakeWebRenderer{},
		OpenAPISpecPath: specPath,
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

func TestNewHandlerSwaggerRedirect(t *testing.T) {
	t.Parallel()

	h := NewHandler(Deps{
		Parser:   &fakeWebParser{},
		Fetcher:  &fakeWebFetcher{},
		Renderer: &fakeWebRenderer{},
	})

	req := httptest.NewRequest(http.MethodGet, "/swagger", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusTemporaryRedirect {
		t.Fatalf("status = %d, want 307", rec.Code)
	}
	if got := rec.Header().Get("Location"); got != "/swagger/index.html" {
		t.Fatalf("redirect location = %q, want /swagger/index.html", got)
	}
}

func TestNewHandlerSwaggerIndexPage(t *testing.T) {
	t.Parallel()

	h := NewHandler(Deps{
		Parser:   &fakeWebParser{},
		Fetcher:  &fakeWebFetcher{},
		Renderer: &fakeWebRenderer{},
	})

	req := httptest.NewRequest(http.MethodGet, "/swagger/index.html", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "Swagger UI") {
		t.Fatalf("swagger index should render swagger ui:\n%s", body)
	}
	if !strings.Contains(body, "/openapi.json") {
		t.Fatalf("swagger index should include local spec endpoint:\n%s", body)
	}
	if !strings.Contains(body, "/swagger/assets/swagger-ui.css") {
		t.Fatalf("swagger index should include local asset paths:\n%s", body)
	}
	if strings.Contains(body, "unpkg.com") {
		t.Fatalf("swagger index should not depend on external CDN:\n%s", body)
	}
}

func TestNewHandlerSwaggerAssetServedLocally(t *testing.T) {
	t.Parallel()

	h := NewHandler(Deps{
		Parser:   &fakeWebParser{},
		Fetcher:  &fakeWebFetcher{},
		Renderer: &fakeWebRenderer{},
	})

	req := httptest.NewRequest(http.MethodGet, "/swagger/assets/swagger-ui.css", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); !strings.Contains(ct, "text/css") {
		t.Fatalf("content-type = %q, want text/css", ct)
	}
}

func TestNewHandlerSwaggerAssetsDoNotDependOnCWD(t *testing.T) {
	t.Parallel()

	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(wd)
	})
	if err := os.Chdir(t.TempDir()); err != nil {
		t.Fatalf("chdir tempdir: %v", err)
	}

	h := NewHandler(Deps{
		Parser:   &fakeWebParser{},
		Fetcher:  &fakeWebFetcher{},
		Renderer: &fakeWebRenderer{},
	})

	reqIndex := httptest.NewRequest(http.MethodGet, "/swagger/index.html", nil)
	recIndex := httptest.NewRecorder()
	h.ServeHTTP(recIndex, reqIndex)
	if recIndex.Code != http.StatusOK {
		t.Fatalf("swagger index status = %d, want 200", recIndex.Code)
	}

	reqAsset := httptest.NewRequest(http.MethodGet, "/swagger/assets/swagger-ui.css", nil)
	recAsset := httptest.NewRecorder()
	h.ServeHTTP(recAsset, reqAsset)
	if recAsset.Code != http.StatusOK {
		t.Fatalf("swagger asset status = %d, want 200", recAsset.Code)
	}
}

func TestNewHandlerConvertStatusMapping(t *testing.T) {
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
		parserErr   error
		fetchErr    error
		renderErr   error
		name        string
		method      string
		body        string
		contentType string
		wantStatus  int
	}{
		{name: "method not allowed", method: http.MethodGet, wantStatus: http.StatusMethodNotAllowed},
		{name: "invalid form", method: http.MethodPost, body: "url=%zz", contentType: "application/x-www-form-urlencoded", wantStatus: http.StatusBadRequest},
		{
			name:        "request body too large",
			method:      http.MethodPost,
			body:        "url=" + strings.Repeat("a", 2*1024*1024),
			contentType: "application/x-www-form-urlencoded",
			wantStatus:  http.StatusRequestEntityTooLarge,
		},
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

			h := NewHandler(Deps{
				Parser:   &fakeWebParser{ref: baseRef, err: tc.parserErr},
				Fetcher:  &fakeWebFetcher{data: baseData, err: tc.fetchErr},
				Renderer: &fakeWebRenderer{content: []byte("# markdown"), err: tc.renderErr},
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

func TestNewHandlerMethodNotAllowedForDocsEndpoints(t *testing.T) {
	t.Parallel()

	h := NewHandler(Deps{
		Parser:   &fakeWebParser{},
		Fetcher:  &fakeWebFetcher{},
		Renderer: &fakeWebRenderer{},
	})

	tcs := []struct {
		path string
	}{
		{path: "/openapi.json"},
		{path: "/swagger"},
		{path: "/swagger/index.html"},
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

func TestFetchHTTPStatusFromError(t *testing.T) {
	t.Parallel()

	tcs := []struct {
		err  error
		name string
		want int
	}{
		{name: "nil", err: nil, want: http.StatusOK},
		{name: "not found", err: gh.ErrResourceNotFound, want: http.StatusNotFound},
		{name: "rate limit", err: errors.New("status 403: API rate limit exceeded"), want: http.StatusTooManyRequests},
		{name: "auth forbidden", err: errors.New("forbidden"), want: http.StatusForbidden},
		{name: "auth unauthorized", err: errors.New("status 401 unauthorized"), want: http.StatusUnauthorized},
		{name: "unknown upstream", err: errors.New("boom"), want: http.StatusBadGateway},
	}

	for _, tc := range tcs {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if got := fetchHTTPStatusFromError(tc.err); got != tc.want {
				t.Fatalf("fetchHTTPStatusFromError(%v) = %d, want %d", tc.err, got, tc.want)
			}
		})
	}
}

func TestFetchStatusFromWrappedStatus(t *testing.T) {
	t.Parallel()

	tcs := []struct {
		err      error
		name     string
		want     int
		wantBool bool
	}{
		{name: "no status", err: errors.New("boom"), want: 0, wantBool: false},
	}

	for _, tc := range tcs {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, ok := fetchStatusFromWrappedStatus(tc.err)
			if got != tc.want || ok != tc.wantBool {
				t.Fatalf("fetchStatusFromWrappedStatus(%v) = (%d,%v), want (%d,%v)", tc.err, got, ok, tc.want, tc.wantBool)
			}
		})
	}
}

func TestAuthHTTPStatus(t *testing.T) {
	t.Parallel()

	tcs := []struct {
		err  error
		name string
		want int
	}{
		{name: "forbidden text", err: errors.New("forbidden"), want: http.StatusForbidden},
		{name: "status 403 text", err: errors.New("http status 403"), want: http.StatusForbidden},
		{name: "fallback unauthorized", err: errors.New("bad credentials"), want: http.StatusUnauthorized},
	}

	for _, tc := range tcs {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := authHTTPStatus(tc.err); got != tc.want {
				t.Fatalf("authHTTPStatus(%v) = %d, want %d", tc.err, got, tc.want)
			}
		})
	}
}

type fakeWebParser struct {
	err error
	ref gh.ResourceRef
}

func (f *fakeWebParser) Parse(rawURL string) (gh.ResourceRef, error) {
	_ = rawURL
	if f.err != nil {
		return gh.ResourceRef{}, f.err
	}
	return f.ref, nil
}

type fakeWebFetcher struct {
	err  error
	data gh.IssueData
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
	err     error
	content []byte
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
