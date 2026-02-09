package config

import (
	"flag"
	"io"
	"os"
)

// Config represents normalized runtime configuration for the CLI.
type Config struct {
	OutputPath      string
	Format          string
	IncludeComments bool
	Stdout          bool
	Force           bool
	InputFile       string
	Positional      []string
	Token           string
	SummaryLang     string
	OpenAIAPIKey    string
	OpenAIBaseURL   string
	OpenAIModel     string
}

// Loader loads configuration from CLI args and environment variables.
type Loader interface {
	Load(args []string) (Config, error)
}

// NewLoader constructs the default configuration loader.
func NewLoader() Loader {
	return &flagLoader{}
}

type flagLoader struct{}

func (l *flagLoader) Load(args []string) (Config, error) {
	_ = l

	cfg := Config{}
	flags := flag.NewFlagSet("issue2md", flag.ContinueOnError)
	flags.SetOutput(io.Discard)

	flags.StringVar(&cfg.OutputPath, "output", "", "output path")
	flags.StringVar(&cfg.Format, "format", "markdown", "output format")
	flags.BoolVar(&cfg.IncludeComments, "include-comments", true, "include comments")
	flags.StringVar(&cfg.InputFile, "input-file", "", "batch input file")
	flags.BoolVar(&cfg.Stdout, "stdout", false, "write markdown to stdout")
	flags.BoolVar(&cfg.Force, "force", false, "overwrite existing files")
	flags.StringVar(&cfg.SummaryLang, "lang", "", "summary language")

	var tokenFlag string
	flags.StringVar(&tokenFlag, "token", "", "GitHub token")

	if err := flags.Parse(args); err != nil {
		return Config{}, WrapError("parse flags", err)
	}

	if cfg.Format != "markdown" {
		return Config{}, WrapError("validate flags", NewValidationError("format", "must be markdown"))
	}
	if cfg.Stdout && cfg.InputFile != "" {
		return Config{}, WrapError("validate flags", NewConflictError("--stdout", "--input-file"))
	}
	cfg.Positional = flags.Args()

	cfg.Token = tokenFlag
	if cfg.Token == "" {
		cfg.Token = os.Getenv("GITHUB_TOKEN")
	}

	cfg.OpenAIAPIKey = os.Getenv("OPENAI_API_KEY")
	cfg.OpenAIBaseURL = os.Getenv("ISSUE2MD_AI_BASE_URL")
	cfg.OpenAIModel = os.Getenv("ISSUE2MD_AI_MODEL")

	return cfg, nil
}
