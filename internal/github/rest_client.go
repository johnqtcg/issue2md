package github

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	goGithub "github.com/google/go-github/v72/github"
	"golang.org/x/oauth2"
)

const defaultRESTBaseURL = "https://api.github.com/"

type restClient struct {
	client *goGithub.Client
}

func newRESTClient(cfg Config) (*restClient, error) {
	httpClient := cfg.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 30 * time.Second}
	}

	if cfg.Token != "" {
		ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: cfg.Token})
		baseTransport := httpClient.Transport
		if baseTransport == nil {
			baseTransport = http.DefaultTransport
		}
		httpClient = &http.Client{
			Transport: &oauth2.Transport{
				Source: ts,
				Base:   baseTransport,
			},
			Timeout: httpClient.Timeout,
		}
	}

	client := goGithub.NewClient(httpClient)

	baseURL := cfg.RESTBaseURL
	if baseURL == "" {
		baseURL = defaultRESTBaseURL
	}
	parsed, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("parse REST base URL %q: %w", baseURL, err)
	}
	client.BaseURL = parsed

	return &restClient{client: client}, nil
}

func (c *restClient) getIssue(ctx context.Context, owner, repo string, number int) (*goGithub.Issue, error) {
	issue, _, err := c.client.Issues.Get(ctx, owner, repo, number)
	if err != nil {
		return nil, wrapRESTError("get issue", err)
	}
	return issue, nil
}

func (c *restClient) listIssueComments(ctx context.Context, owner, repo string, number int) ([]*goGithub.IssueComment, error) {
	var all []*goGithub.IssueComment
	opts := &goGithub.IssueListCommentsOptions{
		ListOptions: goGithub.ListOptions{PerPage: 100},
	}

	for {
		comments, resp, err := c.client.Issues.ListComments(ctx, owner, repo, number, opts)
		if err != nil {
			return nil, wrapRESTError("list issue comments", err)
		}
		all = append(all, comments...)
		if resp == nil || resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return all, nil
}

func (c *restClient) getPullRequest(ctx context.Context, owner, repo string, number int) (*goGithub.PullRequest, error) {
	pr, _, err := c.client.PullRequests.Get(ctx, owner, repo, number)
	if err != nil {
		return nil, wrapRESTError("get pull request", err)
	}
	return pr, nil
}

func (c *restClient) listPullRequestReviews(ctx context.Context, owner, repo string, number int) ([]*goGithub.PullRequestReview, error) {
	var all []*goGithub.PullRequestReview
	opts := &goGithub.ListOptions{PerPage: 100}
	for {
		reviews, resp, err := c.client.PullRequests.ListReviews(ctx, owner, repo, number, opts)
		if err != nil {
			return nil, wrapRESTError("list pull request reviews", err)
		}
		all = append(all, reviews...)
		if resp == nil || resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}
	return all, nil
}

func (c *restClient) listPullRequestComments(ctx context.Context, owner, repo string, number int) ([]*goGithub.PullRequestComment, error) {
	var all []*goGithub.PullRequestComment
	opts := &goGithub.PullRequestListCommentsOptions{
		ListOptions: goGithub.ListOptions{PerPage: 100},
	}
	for {
		comments, resp, err := c.client.PullRequests.ListComments(ctx, owner, repo, number, opts)
		if err != nil {
			return nil, wrapRESTError("list pull request comments", err)
		}
		all = append(all, comments...)
		if resp == nil || resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}
	return all, nil
}

func wrapRESTError(op string, err error) error {
	if err == nil {
		return nil
	}

	var respErr *goGithub.ErrorResponse
	if errors.As(err, &respErr) && respErr.Response != nil {
		return fmt.Errorf("%s: %w", op, &statusError{
			StatusCode: respErr.Response.StatusCode,
			Err:        err,
		})
	}

	return fmt.Errorf("%s: %w", op, err)
}
