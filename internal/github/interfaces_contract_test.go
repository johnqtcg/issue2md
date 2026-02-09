package github

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestConfigWithDefaults(t *testing.T) {
	t.Parallel()

	cfg := Config{}.WithDefaults()

	if cfg.MaxRetries != 3 {
		t.Fatalf("MaxRetries = %d, want 3", cfg.MaxRetries)
	}
	if cfg.InitialBackoff != 2*time.Second {
		t.Fatalf("InitialBackoff = %s, want 2s", cfg.InitialBackoff)
	}
}

func TestNewFetcherReturnsFetcher(t *testing.T) {
	t.Parallel()

	fetcher, err := NewFetcher(Config{})
	if err != nil {
		t.Fatalf("NewFetcher error = %v, want nil", err)
	}
	if fetcher == nil {
		t.Fatal("NewFetcher returned nil fetcher")
	}
}

func TestDefaultFetcherReturnsUnsupportedResourceType(t *testing.T) {
	t.Parallel()

	fetcher, err := NewFetcher(Config{})
	if err != nil {
		t.Fatalf("NewFetcher error = %v, want nil", err)
	}

	_, err = fetcher.Fetch(context.Background(), ResourceRef{}, FetchOptions{})
	if !errors.Is(err, ErrUnsupportedResourceType) {
		t.Fatalf("Fetch error = %v, want ErrUnsupportedResourceType", err)
	}
}
