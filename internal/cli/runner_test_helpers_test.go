package cli

import (
	"context"
	"errors"

	"github.com/johnqtcg/issue2md/internal/config"
	"github.com/johnqtcg/issue2md/internal/converter"
	gh "github.com/johnqtcg/issue2md/internal/github"
)

type fakeLoader struct {
	cfg     config.Config
	err     error
	gotArgs []string
}

func (f *fakeLoader) Load(args []string) (config.Config, error) {
	f.gotArgs = append([]string(nil), args...)
	if f.err != nil {
		return config.Config{}, f.err
	}
	return f.cfg, nil
}

type fakeParser struct {
	refByURL map[string]gh.ResourceRef
	errByURL map[string]error
	gotURLs  []string
}

func (f *fakeParser) Parse(rawURL string) (gh.ResourceRef, error) {
	f.gotURLs = append(f.gotURLs, rawURL)
	if err := f.errByURL[rawURL]; err != nil {
		return gh.ResourceRef{}, err
	}
	ref, ok := f.refByURL[rawURL]
	if !ok {
		return gh.ResourceRef{}, errors.New("unexpected URL")
	}
	return ref, nil
}

type fakeFetcherFactory struct {
	fetcher *fakeFetcher
	err     error
}

func (f *fakeFetcherFactory) New(cfg config.Config) (gh.Fetcher, error) {
	_ = cfg
	if f.err != nil {
		return nil, f.err
	}
	return f.fetcher, nil
}

type fakeFetcher struct {
	dataByURL map[string]gh.IssueData
	errByURL  map[string]error
	gotRefs   []gh.ResourceRef
	gotOpts   []gh.FetchOptions
}

func (f *fakeFetcher) Fetch(_ context.Context, ref gh.ResourceRef, opts gh.FetchOptions) (gh.IssueData, error) {
	f.gotRefs = append(f.gotRefs, ref)
	f.gotOpts = append(f.gotOpts, opts)
	if err := f.errByURL[ref.URL]; err != nil {
		return gh.IssueData{}, err
	}
	data, ok := f.dataByURL[ref.URL]
	if !ok {
		return gh.IssueData{}, errors.New("unexpected ref")
	}
	return data, nil
}

type fakeRendererFactory struct {
	renderer *fakeRenderer
}

func (f *fakeRendererFactory) New(cfg config.Config) converter.Renderer {
	_ = cfg
	return f.renderer
}

type fakeRenderer struct {
	out        []byte
	errByTitle map[string]error
	gotData    []gh.IssueData
	gotOpts    []converter.RenderOptions
}

func (f *fakeRenderer) Render(_ context.Context, data gh.IssueData, opts converter.RenderOptions) ([]byte, error) {
	f.gotData = append(f.gotData, data)
	f.gotOpts = append(f.gotOpts, opts)
	if err := f.errByTitle[data.Meta.Title]; err != nil {
		return nil, err
	}
	return f.out, nil
}

type fakeOutputWriter struct {
	path     string
	errByURL map[string]error
	gotRefs  []gh.ResourceRef
	gotMode  []Mode
}

func (f *fakeOutputWriter) Write(cfg config.Config, mode Mode, ref gh.ResourceRef, markdown []byte) (string, error) {
	_ = cfg
	_ = markdown
	f.gotRefs = append(f.gotRefs, ref)
	f.gotMode = append(f.gotMode, mode)
	if err := f.errByURL[ref.URL]; err != nil {
		return "", err
	}
	if f.path == "" {
		return outputPathStdout, nil
	}
	return f.path, nil
}

type fakeInputReader struct {
	lines   []string
	err     error
	gotPath string
}

func (f *fakeInputReader) Read(path string, handle func(line string) error) error {
	f.gotPath = path
	if f.err != nil {
		return f.err
	}
	for _, line := range f.lines {
		if err := handle(line); err != nil {
			return err
		}
	}
	return nil
}
