package cli

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/johnqtcg/issue2md/internal/config"
	gh "github.com/johnqtcg/issue2md/internal/github"
)

func TestAppRunBatchContinueOnErrorAndSummary(t *testing.T) {
	t.Parallel()

	u1 := "https://github.com/octo/repo/issues/1"
	u2 := "https://github.com/octo/repo/pull/2"
	u3 := "https://github.com/octo/repo/discussions/3"

	loader := &fakeLoader{
		cfg: config.Config{
			InputFile:       "urls.txt",
			OutputPath:      "out",
			IncludeComments: true,
		},
	}
	parser := &fakeParser{
		refByURL: map[string]gh.ResourceRef{
			u1: {Owner: "octo", Repo: "repo", Number: 1, Type: gh.ResourceIssue, URL: u1},
			u3: {Owner: "octo", Repo: "repo", Number: 3, Type: gh.ResourceDiscussion, URL: u3},
		},
		errByURL: map[string]error{u2: errors.New("bad url")},
	}
	fetcher := &fakeFetcher{
		dataByURL: map[string]gh.IssueData{
			u1: minimalIssueData(gh.ResourceIssue, "i1", u1),
			u3: minimalIssueData(gh.ResourceDiscussion, "d3", u3),
		},
		errByURL: map[string]error{},
	}
	renderer := &fakeRenderer{out: []byte("# ok"), errByTitle: map[string]error{}}
	writer := &fakeOutputWriter{path: "out.md", errByURL: map[string]error{}}
	reader := &fakeInputReader{lines: []string{u1, u2, u3}}
	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)

	app := NewApp(AppDeps{
		Loader:          loader,
		Parser:          parser,
		FetcherFactory:  &fakeFetcherFactory{fetcher: fetcher},
		RendererFactory: &fakeRendererFactory{renderer: renderer},
		Writer:          writer,
		InputReader:     reader,
		Stdout:          stdout,
		Stderr:          stderr,
	})

	code := app.Run(context.Background(), []string{"--input-file", "urls.txt", "--output", "out"})
	if code != ExitPartialSuccess {
		t.Fatalf("Run exit code = %d, want %d", code, ExitPartialSuccess)
	}
	if len(fetcher.gotRefs) != 2 {
		t.Fatalf("fetch count = %d, want 2 (continue after one failure)", len(fetcher.gotRefs))
	}
	if reader.gotPath != "urls.txt" {
		t.Fatalf("input path = %q, want urls.txt", reader.gotPath)
	}
	report := stdout.String()
	expected := []string{
		"OK total=3 succeeded=2 failed=1",
		"FAILED url=" + u2 + " type= reason=parse URL",
	}
	for _, piece := range expected {
		if !strings.Contains(report, piece) {
			t.Fatalf("summary missing %q\n%s", piece, report)
		}
	}
}

func TestAppRunBatchAllSuccess(t *testing.T) {
	t.Parallel()

	u1 := "https://github.com/octo/repo/issues/1"
	u2 := "https://github.com/octo/repo/pull/2"

	app := NewApp(AppDeps{
		Loader: &fakeLoader{
			cfg: config.Config{
				InputFile:       "urls.txt",
				OutputPath:      "out",
				IncludeComments: true,
			},
		},
		Parser: &fakeParser{
			refByURL: map[string]gh.ResourceRef{
				u1: {Owner: "octo", Repo: "repo", Number: 1, Type: gh.ResourceIssue, URL: u1},
				u2: {Owner: "octo", Repo: "repo", Number: 2, Type: gh.ResourcePullRequest, URL: u2},
			},
			errByURL: map[string]error{},
		},
		FetcherFactory: &fakeFetcherFactory{
			fetcher: &fakeFetcher{
				dataByURL: map[string]gh.IssueData{
					u1: minimalIssueData(gh.ResourceIssue, "i1", u1),
					u2: minimalIssueData(gh.ResourcePullRequest, "p2", u2),
				},
				errByURL: map[string]error{},
			},
		},
		RendererFactory: &fakeRendererFactory{renderer: &fakeRenderer{out: []byte("# ok"), errByTitle: map[string]error{}}},
		Writer:          &fakeOutputWriter{path: "out.md", errByURL: map[string]error{}},
		InputReader:     &fakeInputReader{lines: []string{u1, u2}},
		Stdout:          new(bytes.Buffer),
		Stderr:          new(bytes.Buffer),
	})

	code := app.Run(context.Background(), []string{"--input-file", "urls.txt", "--output", "out"})
	if code != ExitOK {
		t.Fatalf("Run exit code = %d, want %d", code, ExitOK)
	}
}

func minimalIssueData(resourceType gh.ResourceType, title, url string) gh.IssueData {
	return gh.IssueData{
		Meta: gh.Metadata{
			Type:      resourceType,
			Title:     title,
			Number:    1,
			State:     "open",
			Author:    "alice",
			CreatedAt: "2026-01-01T00:00:00Z",
			UpdatedAt: "2026-01-01T00:00:00Z",
			URL:       url,
		},
		Description: "desc",
	}
}
