package parser

import (
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	gh "github.com/johnqtcg/issue2md/internal/github"
)

// ErrInvalidGitHubURL indicates an input URL is not a supported GitHub resource URL.
var ErrInvalidGitHubURL = errors.New("invalid GitHub URL")

// URLParser parses a raw GitHub URL into a normalized resource reference.
type URLParser interface {
	Parse(rawURL string) (gh.ResourceRef, error)
}

// New creates the default URL parser implementation.
func New() URLParser {
	return &defaultParser{}
}

type defaultParser struct{}

func (p *defaultParser) Parse(rawURL string) (gh.ResourceRef, error) {
	_ = p

	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return gh.ResourceRef{}, fmt.Errorf("parse URL %q: %w", rawURL, err)
	}

	host := strings.ToLower(parsedURL.Hostname())
	if host != "github.com" && host != "www.github.com" {
		return gh.ResourceRef{}, fmt.Errorf("validate URL host %q: %w", host, invalid("unsupported host"))
	}

	owner, repo, kind, number, err := splitAndValidatePath(parsedURL.Path)
	if err != nil {
		return gh.ResourceRef{}, fmt.Errorf("parse URL path %q: %w", parsedURL.Path, err)
	}

	resourceType, err := resolveResourceType(kind)
	if err != nil {
		return gh.ResourceRef{}, fmt.Errorf("resolve resource kind %q: %w", kind, err)
	}

	canonicalURL := fmt.Sprintf("https://github.com/%s/%s/%s/%d", owner, repo, kind, number)

	return gh.ResourceRef{
		Owner:  owner,
		Repo:   repo,
		Number: number,
		Type:   resourceType,
		URL:    canonicalURL,
	}, nil
}

func splitAndValidatePath(rawPath string) (owner, repo, kind string, number int, err error) {
	segments := splitPathSegments(rawPath)
	if len(segments) != 4 {
		return "", "", "", 0, fmt.Errorf("validate path segments: %w", invalid("path must be /{owner}/{repo}/{kind}/{number}"))
	}

	owner = segments[0]
	repo = segments[1]
	kind = segments[2]
	numberText := segments[3]

	if owner == "" || repo == "" {
		return "", "", "", 0, fmt.Errorf("validate owner/repo: %w", invalid("owner/repo must not be empty"))
	}

	number, parseErr := strconv.Atoi(numberText)
	if parseErr != nil || number <= 0 {
		return "", "", "", 0, fmt.Errorf("validate resource number %q: %w", numberText, invalid("resource number must be a positive integer"))
	}

	return owner, repo, kind, number, nil
}

func splitPathSegments(rawPath string) []string {
	trimmed := strings.Trim(rawPath, "/")
	if trimmed == "" {
		return nil
	}
	return strings.Split(trimmed, "/")
}

func resolveResourceType(kind string) (gh.ResourceType, error) {
	switch kind {
	case "issues":
		return gh.ResourceIssue, nil
	case "pull":
		return gh.ResourcePullRequest, nil
	case "discussions":
		return gh.ResourceDiscussion, nil
	default:
		return "", fmt.Errorf("validate resource kind %q: %w", kind, invalid("unsupported resource kind"))
	}
}

func invalid(reason string) error {
	return fmt.Errorf("%w: %s", ErrInvalidGitHubURL, reason)
}
