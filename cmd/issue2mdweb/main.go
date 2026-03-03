package main

// @title issue2md Web API
// @version 1.0
// @description Convert a GitHub Issue/Pull Request/Discussion URL into Markdown.
// @BasePath /

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/johnqtcg/issue2md/internal/config"
	"github.com/johnqtcg/issue2md/internal/converter"
	gh "github.com/johnqtcg/issue2md/internal/github"
	"github.com/johnqtcg/issue2md/internal/parser"
)

const defaultShutdownTimeout = 10 * time.Second

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)

	cfg, err := config.NewLoader().Load(nil)
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	fetcher, err := gh.NewFetcher(gh.Config{
		Token: cfg.Token,
	})
	if err != nil {
		log.Fatalf("create fetcher: %v", err)
	}

	var summarizer converter.Summarizer
	if cfg.OpenAIAPIKey != "" {
		summarizer = converter.NewOpenAISummarizer(converter.OpenAISummarizerConfig{
			APIKey:  cfg.OpenAIAPIKey,
			BaseURL: cfg.OpenAIBaseURL,
			Model:   cfg.OpenAIModel,
		})
	}

	tmpl, err := loadTemplate()
	if err != nil {
		log.Fatalf("load template: %v", err)
	}

	handler := newWebHandler(webDeps{
		parser:          parser.New(),
		fetcher:         fetcher,
		renderer:        converter.NewRenderer(summarizer),
		tmpl:            tmpl,
		openAPISpecPath: defaultOpenAPISpecPath,
	})

	server := &http.Server{
		Addr:    resolveWebAddr(),
		Handler: handler,
	}

	if _, err := fmt.Fprintf(os.Stdout, "issue2md web listening on %s\n", server.Addr); err != nil {
		log.Printf("write startup message: %v", err)
	}
	err = runWithGracefulShutdown(ctx, server.ListenAndServe, server.Shutdown, defaultShutdownTimeout)
	stop()
	if err != nil {
		log.Fatalf("%v", err)
	}
}

func runWithGracefulShutdown(ctx context.Context, serve func() error, shutdown func(context.Context) error, shutdownTimeout time.Duration) error {
	if serve == nil {
		return fmt.Errorf("serve function is nil")
	}
	if shutdown == nil {
		return fmt.Errorf("shutdown function is nil")
	}
	if shutdownTimeout <= 0 {
		shutdownTimeout = defaultShutdownTimeout
	}

	serveErrCh := make(chan error, 1)
	go func() {
		serveErrCh <- serve()
	}()

	select {
	case err := <-serveErrCh:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("serve http: %w", err)
		}
		return nil
	case <-ctx.Done():
		log.Printf("shutdown signal received, stopping web server")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()

		if err := shutdown(shutdownCtx); err != nil && !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("shutdown server: %w", err)
		}

		err := <-serveErrCh
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("serve http: %w", err)
		}
		return nil
	}
}

func resolveWebAddr() string {
	if addr := strings.TrimSpace(os.Getenv("ISSUE2MD_WEB_ADDR")); addr != "" {
		return addr
	}
	return ":8080"
}
