package issue2md_test

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// goVersionFromGoMod parses the canonical Go toolchain version from go.mod.
// It extracts the version on the "go X.Y.Z" line, so callers never need to
// hard-code the version — updating go.mod is the single source of truth.
func goVersionFromGoMod(t *testing.T) string {
	t.Helper()

	data, err := os.ReadFile("go.mod")
	if err != nil {
		t.Fatalf("ReadFile(go.mod): %v", err)
	}

	re := regexp.MustCompile(`(?m)^go\s+(\S+)`)
	m := re.FindSubmatch(data)
	if m == nil {
		t.Fatal("go.mod: could not find 'go X.Y.Z' directive")
	}

	return string(m[1])
}

// TestGoVersionPinnedConsistently verifies that every place the Go toolchain
// version appears (Dockerfile, READMEs) matches the canonical version in
// go.mod.  When Dependabot bumps go.mod, this test will fail fast on the PR,
// signalling that the other files need updating too.
func TestGoVersionPinnedConsistently(t *testing.T) {
	t.Parallel()

	version := goVersionFromGoMod(t)

	testCases := []struct {
		name     string
		path     string
		required []string
	}{
		{
			name:     "Dockerfile builds with go.mod toolchain version",
			path:     "Dockerfile",
			required: []string{"ARG GO_VERSION=" + version},
		},
		{
			name: "README advertises go.mod toolchain version",
			path: "README.md",
			required: []string{
				"go-" + version,
				"go " + version,
			},
		},
		{
			name: "Chinese README advertises go.mod toolchain version",
			path: "README.zh-CN.md",
			required: []string{
				"go-" + version,
				"go " + version,
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			content, err := os.ReadFile(filepath.Clean(tc.path))
			if err != nil {
				t.Fatalf("ReadFile(%q): %v", tc.path, err)
			}

			text := string(content)

			for _, needle := range tc.required {
				if !strings.Contains(text, needle) {
					t.Errorf("%s does not contain %q (go.mod version: %s)", tc.path, needle, version)
				}
			}
		})
	}
}
