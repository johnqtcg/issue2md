package github

import (
	"context"
	"fmt"
)

type fetcher struct {
	cfg  Config
	rest *restClient
	gql  *graphQLClient
}

func (f *fetcher) Fetch(ctx context.Context, ref ResourceRef, opts FetchOptions) (IssueData, error) {
	var (
		data IssueData
		err  error
	)

	switch ref.Type {
	case ResourceIssue:
		err = doWithRetry(ctx, f.cfg.MaxRetries, f.cfg.InitialBackoff, nil, func() error {
			data, err = f.fetchIssue(ctx, ref, opts)
			return err
		})
		if err != nil {
			return IssueData{}, fmt.Errorf("fetch issue: %w", err)
		}
		return data, nil
	case ResourcePullRequest:
		err = doWithRetry(ctx, f.cfg.MaxRetries, f.cfg.InitialBackoff, nil, func() error {
			data, err = f.fetchPullRequest(ctx, ref, opts)
			return err
		})
		if err != nil {
			return IssueData{}, fmt.Errorf("fetch pull request: %w", err)
		}
		return data, nil
	case ResourceDiscussion:
		err = doWithRetry(ctx, f.cfg.MaxRetries, f.cfg.InitialBackoff, nil, func() error {
			data, err = f.fetchDiscussion(ctx, ref, opts)
			return err
		})
		if err != nil {
			return IssueData{}, fmt.Errorf("fetch discussion: %w", err)
		}
		return data, nil
	default:
		return IssueData{}, fmt.Errorf("dispatch resource type %q: %w", ref.Type, ErrUnsupportedResourceType)
	}
}
