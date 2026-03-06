package main

import (
	"context"
	"errors"
	"fmt"
	"html/template"
	"io"
	"log/slog"
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
	"github.com/johnqtcg/issue2md/internal/webapp"
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

func TestNewWebHandlerSwaggerRedirect(t *testing.T) {
	t.Parallel()

	h := newWebHandler(webDeps{
		parser:   &fakeWebParser{},
		fetcher:  &fakeWebFetcher{},
		renderer: &fakeWebRenderer{},
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

func TestNewWebHandlerSwaggerIndexPage(t *testing.T) {
	t.Parallel()

	h := newWebHandler(webDeps{
		parser:   &fakeWebParser{},
		fetcher:  &fakeWebFetcher{},
		renderer: &fakeWebRenderer{},
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

func TestNewWebHandlerSwaggerAssetServedLocally(t *testing.T) {
	t.Parallel()

	h := newWebHandler(webDeps{
		parser:   &fakeWebParser{},
		fetcher:  &fakeWebFetcher{},
		renderer: &fakeWebRenderer{},
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

func TestNewWebHandlerSwaggerAssetsDoNotDependOnCWD(t *testing.T) {
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

	h := newWebHandler(webDeps{
		parser:   &fakeWebParser{},
		fetcher:  &fakeWebFetcher{},
		renderer: &fakeWebRenderer{},
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
			fetchErr:    gh.NewStatusError(http.StatusUnauthorized, errors.New("bad credentials"), nil),
			wantStatus:  http.StatusUnauthorized,
		},
		{
			name:        "fetch auth forbidden",
			method:      http.MethodPost,
			body:        url.Values{"url": []string{rawURL}}.Encode(),
			contentType: "application/x-www-form-urlencoded",
			fetchErr:    gh.NewStatusError(http.StatusForbidden, errors.New("forbidden"), nil),
			wantStatus:  http.StatusForbidden,
		},
		{
			name:        "fetch rate limit",
			method:      http.MethodPost,
			body:        url.Values{"url": []string{rawURL}}.Encode(),
			contentType: "application/x-www-form-urlencoded",
			fetchErr:    gh.NewStatusError(http.StatusForbidden, errors.New("forbidden"), http.Header{"Retry-After": []string{"5"}}),
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

func TestRunWithGracefulShutdownOnContextCancel(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	serveStarted := make(chan struct{})
	serveDone := make(chan struct{})
	shutdownCalled := make(chan context.Context, 1)
	errCh := make(chan error, 1)

	go func() {
		errCh <- runWithGracefulShutdown(
			ctx,
			func() error {
				close(serveStarted)
				<-serveDone
				return http.ErrServerClosed
			},
			func(shutdownCtx context.Context) error {
				shutdownCalled <- shutdownCtx
				close(serveDone)
				return nil
			},
			2*time.Second,
			newTestLogger(),
		)
	}()

	<-serveStarted
	cancel()

	select {
	case shutdownCtx := <-shutdownCalled:
		if _, ok := shutdownCtx.Deadline(); !ok {
			t.Fatal("shutdown context should have deadline")
		}
	case <-time.After(time.Second):
		t.Fatal("shutdown should be called after cancellation")
	}

	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("runWithGracefulShutdown error = %v, want nil", err)
		}
	case <-time.After(time.Second):
		t.Fatal("runWithGracefulShutdown did not return")
	}
}

func TestRunWithGracefulShutdownReturnsServeError(t *testing.T) {
	t.Parallel()

	err := runWithGracefulShutdown(
		context.Background(),
		func() error { return errors.New("boom") },
		func(context.Context) error { return nil },
		time.Second,
		newTestLogger(),
	)
	if err == nil {
		t.Fatal("runWithGracefulShutdown error = nil, want error")
	}
	if !strings.Contains(err.Error(), "serve http: boom") {
		t.Fatalf("error = %v, want serve error context", err)
	}
}

func TestRunWithGracefulShutdownReturnsShutdownError(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	serveStarted := make(chan struct{})
	serveDone := make(chan struct{})
	errCh := make(chan error, 1)

	go func() {
		errCh <- runWithGracefulShutdown(
			ctx,
			func() error {
				close(serveStarted)
				<-serveDone
				return http.ErrServerClosed
			},
			func(context.Context) error {
				close(serveDone)
				return errors.New("shutdown failed")
			},
			time.Second,
			newTestLogger(),
		)
	}()

	<-serveStarted
	cancel()

	select {
	case err := <-errCh:
		if err == nil {
			t.Fatal("runWithGracefulShutdown error = nil, want error")
		}
		if !strings.Contains(err.Error(), "shutdown server: shutdown failed") {
			t.Fatalf("error = %v, want shutdown error context", err)
		}
	case <-time.After(time.Second):
		t.Fatal("runWithGracefulShutdown did not return")
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

func TestResolveWebWriteTimeout(t *testing.T) {
	tcs := []struct {
		name    string
		env     string
		want    time.Duration
		wantErr bool
	}{
		{name: "default", env: "", want: 120 * time.Second},
		{name: "custom seconds", env: "75s", want: 75 * time.Second},
		{name: "custom minutes", env: "2m", want: 2 * time.Minute},
		{name: "trim spaces", env: " 90s ", want: 90 * time.Second},
		{name: "invalid duration", env: "abc", wantErr: true},
		{name: "non positive duration", env: "0s", wantErr: true},
	}

	for _, tc := range tcs {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv(webWriteTimeoutEnv, tc.env)
			got, err := resolveWebWriteTimeout()
			if tc.wantErr {
				if err == nil {
					t.Fatal("resolveWebWriteTimeout() error = nil, want error")
				}
				return
			}
			if err != nil {
				t.Fatalf("resolveWebWriteTimeout() error = %v, want nil", err)
			}
			if got != tc.want {
				t.Fatalf("resolveWebWriteTimeout() = %s, want %s", got, tc.want)
			}
		})
	}
}

func TestNewHTTPServerUsesSafeTimeoutDefaults(t *testing.T) {
	t.Parallel()

	handler := http.NewServeMux()
	server := newHTTPServer(":8080", handler, 120*time.Second)

	if server.Addr != ":8080" {
		t.Fatalf("server.Addr = %q, want :8080", server.Addr)
	}
	if server.Handler != handler {
		t.Fatal("server.Handler does not match input handler")
	}
	if server.ReadHeaderTimeout != 5*time.Second {
		t.Fatalf("ReadHeaderTimeout = %s, want 5s", server.ReadHeaderTimeout)
	}
	if server.ReadTimeout != 15*time.Second {
		t.Fatalf("ReadTimeout = %s, want 15s", server.ReadTimeout)
	}
	if server.WriteTimeout != 120*time.Second {
		t.Fatalf("WriteTimeout = %s, want 120s", server.WriteTimeout)
	}
	if server.IdleTimeout != 60*time.Second {
		t.Fatalf("IdleTimeout = %s, want 60s", server.IdleTimeout)
	}
}

func TestNewHTTPServerFallsBackToDefaultWriteTimeout(t *testing.T) {
	t.Parallel()

	server := newHTTPServer(":8080", http.NewServeMux(), 0)
	if server.WriteTimeout != 120*time.Second {
		t.Fatalf("WriteTimeout = %s, want 120s", server.WriteTimeout)
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

type webDeps struct {
	parser          *fakeWebParser
	fetcher         *fakeWebFetcher
	renderer        *fakeWebRenderer
	tmpl            *template.Template
	openAPISpecPath string
}

func newWebHandler(deps webDeps) http.Handler {
	return webapp.NewHandler(webapp.Deps{
		Parser:          deps.parser,
		Fetcher:         deps.fetcher,
		Renderer:        deps.renderer,
		Template:        deps.tmpl,
		OpenAPISpecPath: deps.openAPISpecPath,
	})
}

func newTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}
