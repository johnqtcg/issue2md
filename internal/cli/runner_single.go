package cli

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/johnqtcg/issue2md/internal/config"
	"github.com/johnqtcg/issue2md/internal/converter"
	gh "github.com/johnqtcg/issue2md/internal/github"
	"github.com/johnqtcg/issue2md/internal/parser"
)

// Runner executes the CLI application flow.
type Runner interface {
	Run(ctx context.Context, args []string) int
}

// FetcherFactory creates GitHub fetcher instances from runtime config.
type FetcherFactory interface {
	New(cfg config.Config) (gh.Fetcher, error)
}

// RendererFactory creates markdown renderer instances from runtime config.
type RendererFactory interface {
	New(cfg config.Config) converter.Renderer
}

// AppDeps defines dependencies for CLI app construction.
type AppDeps struct {
	Loader          config.Loader
	Parser          parser.URLParser
	FetcherFactory  FetcherFactory
	RendererFactory RendererFactory
	Writer          OutputWriter
	InputReader     InputReader
	Stdout          io.Writer
	Stderr          io.Writer
}

// App orchestrates CLI single and batch workflows.
type App struct {
	loader          config.Loader
	parser          parser.URLParser
	fetcherFactory  FetcherFactory
	rendererFactory RendererFactory
	writer          OutputWriter
	inputReader     InputReader
	stdout          io.Writer
	stderr          io.Writer
}

// NewApp creates a CLI runner with injected dependencies.
func NewApp(deps AppDeps) Runner {
	app := &App{
		loader:          deps.Loader,
		parser:          deps.Parser,
		fetcherFactory:  deps.FetcherFactory,
		rendererFactory: deps.RendererFactory,
		writer:          deps.Writer,
		inputReader:     deps.InputReader,
		stdout:          deps.Stdout,
		stderr:          deps.Stderr,
	}
	app.setDefaults()
	return app
}

func (a *App) setDefaults() {
	if a.loader == nil {
		a.loader = config.NewLoader()
	}
	if a.parser == nil {
		a.parser = parser.New()
	}
	if a.fetcherFactory == nil {
		a.fetcherFactory = defaultFetcherFactory{}
	}
	if a.rendererFactory == nil {
		a.rendererFactory = defaultRendererFactory{}
	}
	if a.writer == nil {
		a.writer = NewOutputWriter(os.Stdout)
	}
	if a.inputReader == nil {
		a.inputReader = NewFileInputReader()
	}
	if a.stdout == nil {
		a.stdout = os.Stdout
	}
	if a.stderr == nil {
		a.stderr = os.Stderr
	}
}

// Run executes the CLI workflow and returns an exit code.
func (a *App) Run(ctx context.Context, args []string) int {
	cfg, err := a.loader.Load(args)
	if err != nil {
		writeErrorLine(a.stderr, err)
		return ResolveExitCode(err, false, 0)
	}

	validated, err := ValidateArgs(cfg)
	if err != nil {
		writeErrorLine(a.stderr, err)
		return ResolveExitCode(err, false, 0)
	}

	fetcher, err := a.fetcherFactory.New(cfg)
	if err != nil {
		runErr := fmt.Errorf("build fetcher: %w", err)
		writeErrorLine(a.stderr, runErr)
		return ResolveExitCode(runErr, false, 0)
	}

	renderer := a.rendererFactory.New(cfg)
	singleStatusOutput := a.stdout
	if validated.Mode == ModeSingle && cfg.Stdout {
		// Keep stdout pure markdown when --stdout is used in single mode.
		singleStatusOutput = a.stderr
	}

	switch validated.Mode {
	case ModeSingle:
		item, runErr := a.runSingle(ctx, cfg, validated, fetcher, renderer)
		if runErr != nil {
			item.Status = StatusFailed
			item.Reason = runErr.Error()
			writeStatusLine(singleStatusOutput, item)
			return ResolveExitCode(runErr, false, 0)
		}
		writeStatusLine(singleStatusOutput, item)
		return ExitOK
	case ModeBatch:
		summary, runErr := a.runBatch(ctx, cfg, fetcher, renderer)
		if runErr != nil {
			writeErrorLine(a.stderr, runErr)
		}
		if _, writeErr := fmt.Fprintln(a.stdout, FormatSummary(summary)); writeErr != nil {
			writeErrorLine(a.stderr, fmt.Errorf("write summary output: %w", writeErr))
		}
		return ResolveExitCode(runErr, true, summary.Failed)
	default:
		err = fmt.Errorf("unsupported mode %q", validated.Mode)
		writeErrorLine(a.stderr, err)
		return ResolveExitCode(err, false, 0)
	}
}

func (a *App) runSingle(ctx context.Context, cfg config.Config, args Args, fetcher gh.Fetcher, renderer converter.Renderer) (ItemResult, error) {
	item, err := a.processOne(ctx, cfg, ModeSingle, args.URL, fetcher, renderer)
	if err != nil {
		return item, fmt.Errorf("run single URL %q: %w", args.URL, err)
	}
	return item, nil
}

func (a *App) processOne(ctx context.Context, cfg config.Config, mode Mode, rawURL string, fetcher gh.Fetcher, renderer converter.Renderer) (ItemResult, error) {
	item := ItemResult{
		URL:    rawURL,
		Status: StatusFailed,
	}

	ref, err := a.parser.Parse(rawURL)
	if err != nil {
		return item, fmt.Errorf("parse URL: %w", err)
	}
	item.ResourceType = ref.Type

	data, err := fetcher.Fetch(ctx, ref, gh.FetchOptions{IncludeComments: cfg.IncludeComments})
	if err != nil {
		return item, fmt.Errorf("fetch resource: %w", err)
	}

	markdown, err := renderer.Render(ctx, data, converter.RenderOptions{
		IncludeComments: cfg.IncludeComments,
		IncludeSummary:  true,
		Lang:            cfg.SummaryLang,
	})
	if err != nil {
		return item, fmt.Errorf("render markdown: %w", err)
	}

	outputPath, err := a.writer.Write(cfg, mode, ref, markdown)
	if err != nil {
		return item, fmt.Errorf("write output: %w", err)
	}

	item.Status = StatusOK
	item.OutputPath = outputPath
	return item, nil
}

type defaultFetcherFactory struct{}

func (f defaultFetcherFactory) New(cfg config.Config) (gh.Fetcher, error) {
	_ = f
	fetcher, err := gh.NewFetcher(gh.Config{
		Token: cfg.Token,
	})
	if err != nil {
		return nil, fmt.Errorf("create fetcher: %w", err)
	}
	return fetcher, nil
}

type defaultRendererFactory struct{}

func (f defaultRendererFactory) New(cfg config.Config) converter.Renderer {
	_ = f

	var summarizer converter.Summarizer
	if cfg.OpenAIAPIKey != "" {
		summarizer = converter.NewOpenAISummarizer(converter.OpenAISummarizerConfig{
			APIKey:  cfg.OpenAIAPIKey,
			BaseURL: cfg.OpenAIBaseURL,
			Model:   cfg.OpenAIModel,
		})
	}
	return converter.NewRenderer(summarizer)
}

func writeStatusLine(w io.Writer, item ItemResult) {
	switch item.Status {
	case StatusOK:
		if _, err := fmt.Fprintf(w, "OK url=%s type=%s output=%s\n", item.URL, item.ResourceType, item.OutputPath); err != nil {
			return
		}
	default:
		if _, err := fmt.Fprintf(w, "FAILED url=%s type=%s reason=%s\n", item.URL, item.ResourceType, item.Reason); err != nil {
			return
		}
	}
}

func writeErrorLine(w io.Writer, err error) {
	if _, writeErr := fmt.Fprintf(w, "error: %v\n", err); writeErr != nil {
		return
	}
}
