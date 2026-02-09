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

func TestAppRunSingleSuccess(t *testing.T) {
	t.Parallel()

	url := "https://github.com/octo/repo/issues/1"
	ref := gh.ResourceRef{Owner: "octo", Repo: "repo", Number: 1, Type: gh.ResourceIssue, URL: url}
	data := gh.IssueData{
		Meta: gh.Metadata{
			Type:      gh.ResourceIssue,
			Title:     "issue title",
			Number:    1,
			State:     "open",
			Author:    "alice",
			CreatedAt: "2026-01-01T00:00:00Z",
			UpdatedAt: "2026-01-01T00:00:00Z",
			URL:       url,
		},
		Description: "desc",
	}

	loader := &fakeLoader{
		cfg: config.Config{
			Positional:      []string{url},
			IncludeComments: true,
			SummaryLang:     "zh",
		},
	}
	parser := &fakeParser{refByURL: map[string]gh.ResourceRef{url: ref}, errByURL: map[string]error{}}
	fetcher := &fakeFetcher{dataByURL: map[string]gh.IssueData{url: data}, errByURL: map[string]error{}}
	renderer := &fakeRenderer{out: []byte("# markdown"), errByTitle: map[string]error{}}
	writer := &fakeOutputWriter{path: "out.md", errByURL: map[string]error{}}
	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)

	app := NewApp(AppDeps{
		Loader:          loader,
		Parser:          parser,
		FetcherFactory:  &fakeFetcherFactory{fetcher: fetcher},
		RendererFactory: &fakeRendererFactory{renderer: renderer},
		Writer:          writer,
		InputReader:     &fakeInputReader{},
		Stdout:          stdout,
		Stderr:          stderr,
	})

	code := app.Run(context.Background(), []string{"--lang", "zh", url})
	if code != ExitOK {
		t.Fatalf("Run exit code = %d, want %d", code, ExitOK)
	}
	if len(loader.gotArgs) != 3 {
		t.Fatalf("loader got args len = %d, want 3", len(loader.gotArgs))
	}
	if len(parser.gotURLs) != 1 || parser.gotURLs[0] != url {
		t.Fatalf("parser got URLs = %#v, want [%q]", parser.gotURLs, url)
	}
	if len(fetcher.gotRefs) != 1 || fetcher.gotRefs[0].URL != url {
		t.Fatalf("fetcher got refs = %#v, want %q", fetcher.gotRefs, url)
	}
	if len(fetcher.gotOpts) != 1 || !fetcher.gotOpts[0].IncludeComments {
		t.Fatalf("fetcher opts = %#v, want include comments", fetcher.gotOpts)
	}
	if len(renderer.gotOpts) != 1 || renderer.gotOpts[0].Lang != "zh" {
		t.Fatalf("renderer opts = %#v, want lang zh", renderer.gotOpts)
	}
	if !strings.Contains(stdout.String(), "OK url="+url) {
		t.Fatalf("stdout = %q, want success line", stdout.String())
	}
}

func TestAppRunSingleExitCodeMapping(t *testing.T) {
	t.Parallel()

	url := "https://github.com/octo/repo/issues/2"
	ref := gh.ResourceRef{Owner: "octo", Repo: "repo", Number: 2, Type: gh.ResourceIssue, URL: url}
	baseCfg := config.Config{Positional: []string{url}, IncludeComments: true}

	tcs := []struct {
		name     string
		fetchErr error
		writeErr error
		wantCode int
	}{
		{name: "auth error", fetchErr: errors.New("http status 401: bad credentials"), wantCode: ExitAuth},
		{name: "output conflict", writeErr: ErrOutputConflict, wantCode: ExitOutputConflict},
	}

	for _, tc := range tcs {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			loader := &fakeLoader{cfg: baseCfg}
			parser := &fakeParser{refByURL: map[string]gh.ResourceRef{url: ref}, errByURL: map[string]error{}}
			fetcher := &fakeFetcher{
				dataByURL: map[string]gh.IssueData{
					url: {
						Meta: gh.Metadata{
							Type:      gh.ResourceIssue,
							Title:     "issue title",
							Number:    2,
							State:     "open",
							Author:    "alice",
							CreatedAt: "2026-01-01T00:00:00Z",
							UpdatedAt: "2026-01-01T00:00:00Z",
							URL:       url,
						},
						Description: "desc",
					},
				},
				errByURL: map[string]error{url: tc.fetchErr},
			}
			renderer := &fakeRenderer{out: []byte("# markdown"), errByTitle: map[string]error{}}
			writerErrByURL := map[string]error{}
			if tc.writeErr != nil {
				writerErrByURL[url] = tc.writeErr
			}
			writer := &fakeOutputWriter{path: "out.md", errByURL: writerErrByURL}

			app := NewApp(AppDeps{
				Loader:          loader,
				Parser:          parser,
				FetcherFactory:  &fakeFetcherFactory{fetcher: fetcher},
				RendererFactory: &fakeRendererFactory{renderer: renderer},
				Writer:          writer,
				InputReader:     &fakeInputReader{},
				Stdout:          new(bytes.Buffer),
				Stderr:          new(bytes.Buffer),
			})

			code := app.Run(context.Background(), []string{url})
			if code != tc.wantCode {
				t.Fatalf("Run exit code = %d, want %d", code, tc.wantCode)
			}
		})
	}
}
