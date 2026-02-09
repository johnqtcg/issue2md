package main

// @title issue2md Web API
// @version 1.0
// @description Convert a GitHub Issue/Pull Request/Discussion URL into Markdown.
// @BasePath /

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/johnqtcg/issue2md/internal/config"
	"github.com/johnqtcg/issue2md/internal/converter"
	gh "github.com/johnqtcg/issue2md/internal/github"
	"github.com/johnqtcg/issue2md/internal/parser"
)

func main() {
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
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("serve http: %v", err)
	}
}

func resolveWebAddr() string {
	if addr := strings.TrimSpace(os.Getenv("ISSUE2MD_WEB_ADDR")); addr != "" {
		return addr
	}
	return ":8080"
}
