package webapp

import (
	"strings"
	"testing"
)

func TestLoadTemplate(t *testing.T) {
	t.Parallel()

	tmpl, err := LoadTemplate()
	if err != nil {
		t.Fatalf("LoadTemplate() error = %v", err)
	}

	var out strings.Builder
	err = tmpl.Execute(&out, map[string]any{
		"Markdown": "",
		"URL":      "https://github.com/octo/repo/issues/1",
		"Error":    "",
	})
	if err != nil {
		t.Fatalf("template execute error = %v", err)
	}

	html := out.String()
	if !strings.Contains(html, "issue2md Web") {
		t.Fatalf("rendered template missing title:\n%s", html)
	}
	if !strings.Contains(html, `action="/convert"`) {
		t.Fatalf("rendered template missing convert action:\n%s", html)
	}
}
