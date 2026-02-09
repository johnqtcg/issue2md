package github

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"
)

const (
	// DefaultMaxRetries is the default retry count for GitHub API requests.
	DefaultMaxRetries = 3
	// DefaultInitialBackoff is the first retry delay.
	DefaultInitialBackoff = 2 * time.Second
)

// ErrUnsupportedResourceType indicates the resource type is unknown to the fetcher dispatcher.
var ErrUnsupportedResourceType = errors.New("unsupported github resource type")

// ErrResourceNotFound indicates the requested GitHub resource does not exist.
var ErrResourceNotFound = errors.New("github resource not found")

// FetchOptions controls fetch-time behavior.
type FetchOptions struct {
	IncludeComments bool
}

// Fetcher defines the contract for fetching and normalizing one GitHub resource.
type Fetcher interface {
	Fetch(ctx context.Context, ref ResourceRef, opts FetchOptions) (IssueData, error)
}

// Config configures the GitHub fetcher client.
type Config struct {
	Token          string
	HTTPClient     *http.Client
	MaxRetries     int
	InitialBackoff time.Duration
	RESTBaseURL    string
	GraphQLURL     string
}

// WithDefaults fills missing optional values with package defaults.
func (c Config) WithDefaults() Config {
	if c.MaxRetries == 0 {
		c.MaxRetries = DefaultMaxRetries
	}
	if c.InitialBackoff == 0 {
		c.InitialBackoff = DefaultInitialBackoff
	}
	return c
}

// NewFetcher constructs a fetcher instance.
func NewFetcher(cfg Config) (Fetcher, error) {
	cfg = cfg.WithDefaults()
	if cfg.MaxRetries < 0 {
		return nil, fmt.Errorf("invalid MaxRetries %d", cfg.MaxRetries)
	}
	if cfg.InitialBackoff < 0 {
		return nil, fmt.Errorf("invalid InitialBackoff %s", cfg.InitialBackoff)
	}

	restClient, err := newRESTClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("create REST client: %w", err)
	}

	graphQLClient := newGraphQLClient(cfg)

	return &fetcher{
		cfg:  cfg,
		rest: restClient,
		gql:  graphQLClient,
	}, nil
}
