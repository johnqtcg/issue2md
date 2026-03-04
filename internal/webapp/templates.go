package webapp

import (
	"fmt"
	"html/template"
	"io/fs"

	webassets "github.com/johnqtcg/issue2md/web"
)

// LoadTemplate loads the HTML template from embedded assets with fallback support.
func LoadTemplate() (*template.Template, error) {
	return loadTemplateFromFS(webassets.FS, "templates/index.html", DefaultIndexTemplate)
}

func loadTemplateFromFS(assets fs.FS, pattern, fallbackTemplate string) (*template.Template, error) {
	tmpl, err := template.ParseFS(assets, pattern)
	if err == nil {
		return tmpl, nil
	}

	fallback, fallbackErr := template.New("index").Parse(fallbackTemplate)
	if fallbackErr != nil {
		return nil, fmt.Errorf("parse fallback template after loading %q failed (%v): %w", pattern, err, fallbackErr)
	}
	return fallback, nil
}

// DefaultIndexTemplate is the embedded fallback index template.
const DefaultIndexTemplate = `<!doctype html>
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
