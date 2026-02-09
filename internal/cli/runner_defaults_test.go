package cli

import (
	"bytes"
	"context"
	"testing"

	"github.com/johnqtcg/issue2md/internal/config"
)

func TestAppRunWithDefaultDepsHandlesLoaderError(t *testing.T) {
	t.Parallel()

	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	app := NewApp(AppDeps{
		Stdout: stdout,
		Stderr: stderr,
	})

	code := app.Run(context.Background(), []string{"--format", "json"})
	if code != ExitInvalidArguments {
		t.Fatalf("Run exit code = %d, want %d", code, ExitInvalidArguments)
	}
	if stderr.Len() == 0 {
		t.Fatal("stderr should contain error output")
	}
}

func TestDefaultFactoriesConstructors(t *testing.T) {
	t.Parallel()

	fetcherFactory := defaultFetcherFactory{}
	fetcher, err := fetcherFactory.New(config.Config{})
	if err != nil {
		t.Fatalf("defaultFetcherFactory.New error = %v, want nil", err)
	}
	if fetcher == nil {
		t.Fatal("defaultFetcherFactory.New returned nil fetcher")
	}

	rendererFactory := defaultRendererFactory{}
	renderer := rendererFactory.New(config.Config{})
	if renderer == nil {
		t.Fatal("defaultRendererFactory.New returned nil renderer")
	}

	rendererWithSummary := rendererFactory.New(config.Config{OpenAIAPIKey: "k", OpenAIModel: "gpt-5-mini"})
	if rendererWithSummary == nil {
		t.Fatal("defaultRendererFactory.New with API key returned nil renderer")
	}
}
