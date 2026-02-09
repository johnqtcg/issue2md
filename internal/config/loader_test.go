package config

import (
	"errors"
	"testing"
)

func TestLoaderTokenPriority(t *testing.T) {
	t.Setenv("GITHUB_TOKEN", "env-token")

	loader := NewLoader()
	cfg, err := loader.Load([]string{"--token", "flag-token"})
	if err != nil {
		t.Fatalf("Load error = %v, want nil", err)
	}
	if cfg.Token != "flag-token" {
		t.Fatalf("Token = %q, want flag-token", cfg.Token)
	}
}

func TestLoaderTokenFallbackToEnv(t *testing.T) {
	t.Setenv("GITHUB_TOKEN", "env-token")

	loader := NewLoader()
	cfg, err := loader.Load(nil)
	if err != nil {
		t.Fatalf("Load error = %v, want nil", err)
	}
	if cfg.Token != "env-token" {
		t.Fatalf("Token = %q, want env-token", cfg.Token)
	}
}

func TestLoaderRejectsInvalidFormat(t *testing.T) {
	t.Parallel()

	loader := NewLoader()
	_, err := loader.Load([]string{"--format", "json"})
	if err == nil {
		t.Fatal("Load error = nil, want error")
	}

	var vErr *ValidationError
	if !errors.As(err, &vErr) {
		t.Fatalf("Load error = %T, want *ValidationError", err)
	}
	if vErr.Field != "format" {
		t.Fatalf("ValidationError.Field = %q, want format", vErr.Field)
	}
	if err.Error() == vErr.Error() {
		t.Fatalf("Load error = %q, want wrapped error context", err)
	}
}

func TestLoaderLoadsLang(t *testing.T) {
	t.Parallel()

	loader := NewLoader()
	cfg, err := loader.Load([]string{"--lang", "zh"})
	if err != nil {
		t.Fatalf("Load error = %v, want nil", err)
	}
	if cfg.SummaryLang != "zh" {
		t.Fatalf("SummaryLang = %q, want zh", cfg.SummaryLang)
	}
}

func TestLoaderRejectsStdoutWithInputFile(t *testing.T) {
	t.Parallel()

	loader := NewLoader()
	_, err := loader.Load([]string{"--stdout", "--input-file", "urls.txt"})
	if err == nil {
		t.Fatal("Load error = nil, want error")
	}

	var cErr *ConflictError
	if !errors.As(err, &cErr) {
		t.Fatalf("Load error = %T, want *ConflictError", err)
	}
	if err.Error() == cErr.Error() {
		t.Fatalf("Load error = %q, want wrapped error context", err)
	}
}

func TestLoaderStoresPositionalArgs(t *testing.T) {
	t.Parallel()

	loader := NewLoader()
	cfg, err := loader.Load([]string{"--lang", "en", "https://github.com/octo/repo/issues/1"})
	if err != nil {
		t.Fatalf("Load error = %v, want nil", err)
	}
	if len(cfg.Positional) != 1 {
		t.Fatalf("Positional len = %d, want 1", len(cfg.Positional))
	}
	if cfg.Positional[0] != "https://github.com/octo/repo/issues/1" {
		t.Fatalf("Positional[0] = %q, want issue URL", cfg.Positional[0])
	}
}
