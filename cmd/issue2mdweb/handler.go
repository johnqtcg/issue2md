package main

import (
	"html/template"
	"net/http"

	"github.com/johnqtcg/issue2md/internal/converter"
	gh "github.com/johnqtcg/issue2md/internal/github"
	"github.com/johnqtcg/issue2md/internal/parser"
	"github.com/johnqtcg/issue2md/internal/webapp"
)

const defaultOpenAPISpecPath = webapp.DefaultOpenAPISpecPath

type webDeps struct {
	parser          parser.URLParser
	fetcher         gh.Fetcher
	renderer        converter.Renderer
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
