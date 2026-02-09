package main

import (
	"fmt"
	"html/template"
	"os"
	"path/filepath"
)

func loadTemplate() (*template.Template, error) {
	path := filepath.Join("web", "templates", "index.html")
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return template.New("index").Parse(defaultIndexTemplate)
		}
		return nil, fmt.Errorf("stat template file: %w", err)
	}

	tmpl, err := template.ParseFiles(path)
	if err != nil {
		return nil, fmt.Errorf("parse template file: %w", err)
	}
	return tmpl, nil
}

const defaultIndexTemplate = `<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>issue2md Web</title>
  <link rel="stylesheet" href="/static/style.css">
</head>
<body>
  <main class="container">
    <h1>issue2md Web</h1>
    <p><a href="/swagger" target="_blank" rel="noreferrer">Open API docs</a></p>
    <form method="post" action="/convert">
      <label for="url">GitHub URL</label>
      <input id="url" name="url" type="url" required value="{{ .URL }}">
      <button type="submit">Convert</button>
    </form>
    {{ if .Error }}<p class="error">{{ .Error }}</p>{{ end }}
    {{ if .Markdown }}<pre>{{ .Markdown }}</pre>{{ end }}
  </main>
</body>
</html>`

const swaggerDocsPage = `<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>issue2md OpenAPI</title>
</head>
<body>
  <main>
    <h1>issue2md OpenAPI</h1>
    <p>OpenAPI spec is generated locally by <code>make swagger</code>.</p>
    <p><a href="/openapi.json">/openapi.json</a></p>
  </main>
</body>
</html>`
