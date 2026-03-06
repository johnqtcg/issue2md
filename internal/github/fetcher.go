package github

import (
	"context"
	"fmt"
)

type fetcher struct {
	rest *restClient
	gql  *graphQLClient
	cfg  Config
}

func (f *fetcher) Fetch(ctx context.Context, ref ResourceRef, opts FetchOptions) (IssueData, error) {
	switch ref.Type {
	case ResourceIssue:
		return f.fetchWithRetry(ctx, "issue", func() (IssueData, error) {
			return f.fetchIssue(ctx, ref, opts)
		})
	case ResourcePullRequest:
		return f.fetchWithRetry(ctx, "pull request", func() (IssueData, error) {
			return f.fetchPullRequest(ctx, ref, opts)
		})
	case ResourceDiscussion:
		return f.fetchWithRetry(ctx, "discussion", func() (IssueData, error) {
			return f.fetchDiscussion(ctx, ref, opts)
		})
	default:
		return IssueData{}, fmt.Errorf("dispatch resource type %q: %w", ref.Type, ErrUnsupportedResourceType)
	}
}

func (f *fetcher) fetchWithRetry(ctx context.Context, label string, fn func() (IssueData, error)) (IssueData, error) {
	var data IssueData

	err := doWithRetry(ctx, f.cfg.MaxRetries, f.cfg.InitialBackoff, nil, func() error {
		var err error
		data, err = fn()
		return err
	})
	if err != nil {
		return IssueData{}, fmt.Errorf("fetch %s: %w", label, err)
	}
	return data, nil
}
