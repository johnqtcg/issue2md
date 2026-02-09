package main

import (
	"errors"
	"html/template"
	"net/http"
	"os"
	"strings"

	"github.com/johnqtcg/issue2md/internal/converter"
	gh "github.com/johnqtcg/issue2md/internal/github"
	"github.com/johnqtcg/issue2md/internal/parser"
)

const defaultOpenAPISpecPath = "docs/swagger.json"

type webDeps struct {
	parser          parser.URLParser
	fetcher         gh.Fetcher
	renderer        converter.Renderer
	tmpl            *template.Template
	openAPISpecPath string
}

type webHandler struct {
	parser          parser.URLParser
	fetcher         gh.Fetcher
	renderer        converter.Renderer
	tmpl            *template.Template
	openAPISpecPath string
}

func newWebHandler(deps webDeps) http.Handler {
	tmpl := deps.tmpl
	if tmpl == nil {
		tmpl = template.Must(template.New("index").Parse(defaultIndexTemplate))
	}

	openAPISpecPath := deps.openAPISpecPath
	if openAPISpecPath == "" {
		openAPISpecPath = defaultOpenAPISpecPath
	}

	handler := &webHandler{
		parser:          deps.parser,
		fetcher:         deps.fetcher,
		renderer:        deps.renderer,
		tmpl:            tmpl,
		openAPISpecPath: openAPISpecPath,
	}

	mux := http.NewServeMux()
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("web/static"))))
	mux.HandleFunc("/", handler.handleIndex)
	mux.HandleFunc("/convert", handler.handleConvert)
	mux.HandleFunc("/openapi.json", handler.handleOpenAPISpec)
	mux.HandleFunc("/swagger", handler.handleSwaggerUI)

	return mux
}

func (h *webHandler) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if err := h.tmpl.Execute(w, map[string]any{
		"Markdown": "",
		"URL":      "",
		"Error":    "",
	}); err != nil {
		http.Error(w, "render template failed", http.StatusInternalServerError)
	}
}

// handleConvert converts a GitHub resource URL to markdown.
// @Summary Convert GitHub URL to Markdown
// @Description Fetch one GitHub issue, pull request, or discussion and render it as markdown.
// @Tags convert
// @Accept application/x-www-form-urlencoded
// @Produce plain
// @Param url formData string true "GitHub issue/pull/discussion URL"
// @Success 200 {string} string "markdown body"
// @Failure 400 {string} string "invalid request"
// @Failure 401 {string} string "unauthorized"
// @Failure 403 {string} string "forbidden"
// @Failure 404 {string} string "resource not found"
// @Failure 429 {string} string "rate limited"
// @Failure 502 {string} string "upstream failure"
// @Failure 500 {string} string "render failed"
// @Router /convert [post]
func (h *webHandler) handleConvert(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid form", http.StatusBadRequest)
		return
	}

	rawURL := r.FormValue("url")
	if rawURL == "" {
		http.Error(w, "missing url", http.StatusBadRequest)
		return
	}

	ref, err := h.parser.Parse(rawURL)
	if err != nil {
		http.Error(w, "invalid github url", http.StatusBadRequest)
		return
	}

	data, err := h.fetcher.Fetch(r.Context(), ref, gh.FetchOptions{IncludeComments: true})
	if err != nil {
		http.Error(w, "fetch github resource failed", fetchHTTPStatusFromError(err))
		return
	}

	markdown, err := h.renderer.Render(r.Context(), data, converter.RenderOptions{
		IncludeComments: true,
		IncludeSummary:  true,
	})
	if err != nil {
		http.Error(w, "render markdown failed", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	if _, err := w.Write(markdown); err != nil {
		http.Error(w, "write response failed", http.StatusInternalServerError)
	}
}

// handleOpenAPISpec serves generated OpenAPI JSON.
// @Summary Get OpenAPI specification
// @Description Returns generated OpenAPI JSON. Run `make swagger` before calling this endpoint.
// @Tags docs
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 503 {string} string "spec unavailable"
// @Failure 500 {string} string "read failed"
// @Router /openapi.json [get]
func (h *webHandler) handleOpenAPISpec(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	spec, err := os.ReadFile(h.openAPISpecPath)
	if err != nil {
		if os.IsNotExist(err) {
			http.Error(w, "openapi spec not generated, run: make swagger", http.StatusServiceUnavailable)
			return
		}
		http.Error(w, "read openapi spec failed", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if _, err := w.Write(spec); err != nil {
		http.Error(w, "write response failed", http.StatusInternalServerError)
	}
}

func (h *webHandler) handleSwaggerUI(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if _, err := w.Write([]byte(swaggerDocsPage)); err != nil {
		http.Error(w, "write response failed", http.StatusInternalServerError)
	}
}

func fetchHTTPStatusFromError(err error) int {
	if err == nil {
		return http.StatusOK
	}

	if status, ok := fetchStatusFromClassifiedError(err); ok {
		return status
	}
	if status, ok := fetchStatusFromWrappedStatus(err); ok {
		return status
	}

	return http.StatusBadGateway
}

func fetchStatusFromClassifiedError(err error) (int, bool) {
	if errors.Is(err, gh.ErrResourceNotFound) {
		return http.StatusNotFound, true
	}
	if gh.IsRateLimitError(err) {
		return http.StatusTooManyRequests, true
	}
	if gh.IsAuthError(err) {
		return authHTTPStatus(err), true
	}
	return 0, false
}

func authHTTPStatus(err error) int {
	if status, ok := gh.StatusCode(err); ok {
		if status == http.StatusUnauthorized || status == http.StatusForbidden {
			return status
		}
	}

	text := strings.ToLower(err.Error())
	if strings.Contains(text, "status 403") || strings.Contains(text, "forbidden") {
		return http.StatusForbidden
	}
	return http.StatusUnauthorized
}

func fetchStatusFromWrappedStatus(err error) (int, bool) {
	status, ok := gh.StatusCode(err)
	if !ok {
		return 0, false
	}

	switch status {
	case http.StatusNotFound:
		return http.StatusNotFound, true
	case http.StatusTooManyRequests:
		return http.StatusTooManyRequests, true
	case http.StatusUnauthorized, http.StatusForbidden:
		return status, true
	default:
		if status >= 500 && status <= 599 {
			return http.StatusBadGateway, true
		}
		return 0, false
	}

}
