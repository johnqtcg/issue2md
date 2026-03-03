package main

import (
	"html/template"

	"github.com/johnqtcg/issue2md/internal/webapp"
)

func loadTemplate() (*template.Template, error) {
	return webapp.LoadTemplate()
}
