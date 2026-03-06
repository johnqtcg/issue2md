package main

// @title issue2md Web API
// @version 1.0
// @description Convert a GitHub Issue/Pull Request/Discussion URL into Markdown.
// @BasePath /

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
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
	"github.com/johnqtcg/issue2md/internal/webapp"
)

const (
	defaultShutdownTimeout   = 10 * time.Second
	defaultReadHeaderTimeout = 5 * time.Second
	defaultReadTimeout       = 15 * time.Second
	// /convert may spend tens of seconds in upstream fetch/summarization before first byte.
	// Keep write timeout comfortably above known upstream client timeouts (30s/45s).
	defaultWriteTimeout = 120 * time.Second
	defaultIdleTimeout  = 60 * time.Second
)

const webWriteTimeoutEnv = "ISSUE2MD_WEB_WRITE_TIMEOUT"

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	logger := newLogger(os.Stderr)

	cfg, err := config.NewLoader().Load(nil)
	if err != nil {
		fatal(logger, "load config", err)
	}

	fetcher, err := gh.NewFetcher(gh.Config{
		Token: cfg.Token,
	})
	if err != nil {
		fatal(logger, "create fetcher", err)
	}

	var summarizer converter.Summarizer
	if cfg.OpenAIAPIKey != "" {
		summarizer = converter.NewOpenAISummarizer(converter.OpenAISummarizerConfig{
			AuthValue: cfg.OpenAIAPIKey,
			BaseURL:   cfg.OpenAIBaseURL,
			Model:     cfg.OpenAIModel,
		})
	}

	tmpl, err := loadTemplate()
	if err != nil {
		fatal(logger, "load template", err)
	}

	handler := webapp.NewHandler(webapp.Deps{
		Parser:          parser.New(),
		Fetcher:         fetcher,
		Renderer:        converter.NewRenderer(summarizer),
		Template:        tmpl,
		OpenAPISpecPath: webapp.DefaultOpenAPISpecPath,
	})

	writeTimeout, err := resolveWebWriteTimeout()
	if err != nil {
		fatal(logger, "resolve web write timeout", err)
	}

	server := newHTTPServer(resolveWebAddr(), handler, writeTimeout)

	if _, err := fmt.Fprintf(os.Stdout, "issue2md web listening on %s\n", server.Addr); err != nil {
		logger.Error("write startup message", "err", err)
	}
	err = runWithGracefulShutdown(ctx, server.ListenAndServe, server.Shutdown, defaultShutdownTimeout, logger)
	stop()
	if err != nil {
		fatal(logger, "run web server", err)
	}
}

func runWithGracefulShutdown(ctx context.Context, serve func() error, shutdown func(context.Context) error, shutdownTimeout time.Duration, logger *slog.Logger) error {
	if serve == nil {
		return fmt.Errorf("serve function is nil")
	}
	if shutdown == nil {
		return fmt.Errorf("shutdown function is nil")
	}
	if shutdownTimeout <= 0 {
		shutdownTimeout = defaultShutdownTimeout
	}
	if logger == nil {
		logger = newLogger(io.Discard)
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
		logger.Info("shutdown signal received", "timeout", shutdownTimeout.String())

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

func newLogger(w io.Writer) *slog.Logger {
	return slog.New(slog.NewTextHandler(w, nil))
}

func fatal(logger *slog.Logger, msg string, err error) {
	logger.Error(msg, "err", err)
	os.Exit(1)
}

func resolveWebAddr() string {
	if addr := strings.TrimSpace(os.Getenv("ISSUE2MD_WEB_ADDR")); addr != "" {
		return addr
	}
	return ":8080"
}

func resolveWebWriteTimeout() (time.Duration, error) {
	value := strings.TrimSpace(os.Getenv(webWriteTimeoutEnv))
	if value == "" {
		return defaultWriteTimeout, nil
	}

	timeout, err := time.ParseDuration(value)
	if err != nil {
		return 0, fmt.Errorf("%s=%q: parse duration: %w", webWriteTimeoutEnv, value, err)
	}
	if timeout <= 0 {
		return 0, fmt.Errorf("%s=%q: duration must be greater than 0", webWriteTimeoutEnv, value)
	}
	return timeout, nil
}

func newHTTPServer(addr string, handler http.Handler, writeTimeout time.Duration) *http.Server {
	if writeTimeout <= 0 {
		writeTimeout = defaultWriteTimeout
	}

	return &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadHeaderTimeout: defaultReadHeaderTimeout,
		ReadTimeout:       defaultReadTimeout,
		WriteTimeout:      writeTimeout,
		IdleTimeout:       defaultIdleTimeout,
	}
}
