package cli

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/johnqtcg/issue2md/internal/config"
	gh "github.com/johnqtcg/issue2md/internal/github"
)

func TestDefaultFileName(t *testing.T) {
	t.Parallel()

	tcs := []struct {
		name string
		ref  gh.ResourceRef
		want string
	}{
		{
			name: "issue",
			ref:  gh.ResourceRef{Owner: "octo", Repo: "repo", Type: gh.ResourceIssue, Number: 1},
			want: "octo-repo-issue-1.md",
		},
		{
			name: "pr",
			ref:  gh.ResourceRef{Owner: "octo", Repo: "repo", Type: gh.ResourcePullRequest, Number: 2},
			want: "octo-repo-pr-2.md",
		},
		{
			name: "discussion",
			ref:  gh.ResourceRef{Owner: "octo", Repo: "repo", Type: gh.ResourceDiscussion, Number: 3},
			want: "octo-repo-discussion-3.md",
		},
	}

	for _, tc := range tcs {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := defaultFileName(tc.ref)
			if err != nil {
				t.Fatalf("defaultFileName error = %v, want nil", err)
			}
			if got != tc.want {
				t.Fatalf("defaultFileName = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestOutputWriterWritePathBehavior(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	buf := new(bytes.Buffer)
	w := NewOutputWriter(buf)
	ref := gh.ResourceRef{Owner: "octo", Repo: "repo", Type: gh.ResourceIssue, Number: 9}

	t.Run("single mode explicit file", func(t *testing.T) {
		target := filepath.Join(tmpDir, "one.md")
		gotPath, err := w.Write(config.Config{OutputPath: target}, ModeSingle, ref, []byte("hello"))
		if err != nil {
			t.Fatalf("Write error = %v, want nil", err)
		}
		if gotPath != target {
			t.Fatalf("Write path = %q, want %q", gotPath, target)
		}
		content, err := os.ReadFile(target)
		if err != nil {
			t.Fatalf("ReadFile error = %v", err)
		}
		if string(content) != "hello" {
			t.Fatalf("content = %q, want hello", string(content))
		}
	})

	t.Run("single mode output directory", func(t *testing.T) {
		dir := filepath.Join(tmpDir, "out-dir")
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatalf("MkdirAll error = %v", err)
		}

		gotPath, err := w.Write(config.Config{OutputPath: dir}, ModeSingle, ref, []byte("abc"))
		if err != nil {
			t.Fatalf("Write error = %v, want nil", err)
		}
		wantPath := filepath.Join(dir, "octo-repo-issue-9.md")
		if gotPath != wantPath {
			t.Fatalf("Write path = %q, want %q", gotPath, wantPath)
		}
	})

	t.Run("batch mode uses output directory", func(t *testing.T) {
		dir := filepath.Join(tmpDir, "batch")
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatalf("MkdirAll error = %v", err)
		}

		gotPath, err := w.Write(config.Config{OutputPath: dir}, ModeBatch, ref, []byte("xyz"))
		if err != nil {
			t.Fatalf("Write error = %v, want nil", err)
		}
		wantPath := filepath.Join(dir, "octo-repo-issue-9.md")
		if gotPath != wantPath {
			t.Fatalf("Write path = %q, want %q", gotPath, wantPath)
		}
	})
}

func TestOutputWriterStdoutMode(t *testing.T) {
	t.Parallel()

	buf := new(bytes.Buffer)
	w := NewOutputWriter(buf)

	gotPath, err := w.Write(config.Config{Stdout: true}, ModeSingle, gh.ResourceRef{
		Owner: "octo", Repo: "repo", Number: 1, Type: gh.ResourceIssue,
	}, []byte("# title"))
	if err != nil {
		t.Fatalf("Write error = %v, want nil", err)
	}
	if gotPath != outputPathStdout {
		t.Fatalf("Write path = %q, want %q", gotPath, outputPathStdout)
	}
	if buf.String() != "# title" {
		t.Fatalf("stdout = %q, want %q", buf.String(), "# title")
	}
}

func TestOutputWriterForceOverwriteRule(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	target := filepath.Join(tmpDir, "conflict.md")
	if err := os.WriteFile(target, []byte("old"), 0o644); err != nil {
		t.Fatalf("WriteFile setup error = %v", err)
	}

	w := NewOutputWriter(new(bytes.Buffer))
	ref := gh.ResourceRef{Owner: "octo", Repo: "repo", Type: gh.ResourceIssue, Number: 3}

	_, err := w.Write(config.Config{OutputPath: target, Force: false}, ModeSingle, ref, []byte("new"))
	if !errors.Is(err, ErrOutputConflict) {
		t.Fatalf("Write error = %v, want ErrOutputConflict", err)
	}

	_, err = w.Write(config.Config{OutputPath: target, Force: true}, ModeSingle, ref, []byte("new"))
	if err != nil {
		t.Fatalf("Write with force error = %v, want nil", err)
	}
	content, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("ReadFile error = %v", err)
	}
	if string(content) != "new" {
		t.Fatalf("content = %q, want new", string(content))
	}
}
