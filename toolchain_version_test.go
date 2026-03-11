package issue2md_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPatchedGoVersionIsPinnedConsistently(t *testing.T) {
	t.Parallel()

	const patchedVersion = "1.25.8"
	const vulnerableVersion = "1.25.7"

	testCases := []struct {
		name     string
		path     string
		required []string
		forbid   []string
	}{
		{
			name:     "go.mod pins patched toolchain",
			path:     "go.mod",
			required: []string{"go " + patchedVersion},
			forbid:   []string{"go " + vulnerableVersion},
		},
		{
			name:     "Dockerfile builds with patched Go image",
			path:     "Dockerfile",
			required: []string{"ARG GO_VERSION=" + patchedVersion},
			forbid:   []string{"ARG GO_VERSION=" + vulnerableVersion},
		},
		{
			name: "README advertises patched Go version",
			path: "README.md",
			required: []string{
				"go-" + patchedVersion,
				"go " + patchedVersion,
			},
			forbid: []string{
				"go-" + vulnerableVersion,
				"go " + vulnerableVersion,
			},
		},
		{
			name: "Chinese README advertises patched Go version",
			path: "README.zh-CN.md",
			required: []string{
				"go-" + patchedVersion,
				"go " + patchedVersion,
			},
			forbid: []string{
				"go-" + vulnerableVersion,
				"go " + vulnerableVersion,
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			content, err := os.ReadFile(filepath.Clean(tc.path))
			if err != nil {
				t.Fatalf("ReadFile(%q) error = %v", tc.path, err)
			}

			text := string(content)

			for _, needle := range tc.required {
				if !strings.Contains(text, needle) {
					t.Errorf("%s does not contain %q", tc.path, needle)
				}
			}

			for _, needle := range tc.forbid {
				if strings.Contains(text, needle) {
					t.Errorf("%s unexpectedly contains vulnerable marker %q", tc.path, needle)
				}
			}
		})
	}
}
