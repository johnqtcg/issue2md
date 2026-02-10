package github

import (
	"context"
	"net/http"
	"testing"
)

func TestFetchPullRequest(t *testing.T) {
	t.Parallel()

	clientHTTP := newTestHTTPClient(func(r *http.Request) (*http.Response, error) {
		switch r.URL.Path {
		case "/repos/octo/repo/issues/2":
			return mustJSONResponse(t, http.StatusOK, map[string]any{
				"number":     2,
				"title":      "PR as issue",
				"state":      "open",
				"body":       "Issue envelope for reactions",
				"html_url":   "https://github.com/octo/repo/issues/2",
				"created_at": "2026-01-01T00:00:00Z",
				"updated_at": "2026-01-02T00:00:00Z",
				"user":       map[string]any{"login": "alice"},
				"reactions": map[string]any{
					"+1":          3,
					"heart":       1,
					"total_count": 4,
				},
			}), nil
		case "/repos/octo/repo/issues/2/comments":
			return mustJSONResponse(t, http.StatusOK, []map[string]any{
				{
					"id":         1101,
					"body":       "top-level PR conversation comment",
					"html_url":   "https://github.com/octo/repo/pull/2#issuecomment-1101",
					"created_at": "2026-01-03T00:30:00Z",
					"updated_at": "2026-01-03T00:30:00Z",
					"user":       map[string]any{"login": "issue-commenter"},
				},
			}), nil
		case "/repos/octo/repo/pulls/2":
			return mustJSONResponse(t, http.StatusOK, map[string]any{
				"number":          2,
				"title":           "PR title",
				"state":           "open",
				"body":            "PR body",
				"html_url":        "https://github.com/octo/repo/pull/2",
				"created_at":      "2026-01-01T00:00:00Z",
				"updated_at":      "2026-01-02T00:00:00Z",
				"merged":          true,
				"merged_at":       "2026-01-03T00:00:00Z",
				"user":            map[string]any{"login": "alice"},
				"labels":          []map[string]any{{"name": "enhancement"}},
				"review_comments": 1,
			}), nil
		case "/repos/octo/repo/pulls/2/reviews":
			return mustJSONResponse(t, http.StatusOK, []map[string]any{
				{
					"id":           2001,
					"state":        "APPROVED",
					"body":         "looks good",
					"submitted_at": "2026-01-03T01:00:00Z",
					"user":         map[string]any{"login": "reviewer"},
				},
			}), nil
		case "/repos/octo/repo/pulls/2/comments":
			return mustJSONResponse(t, http.StatusOK, []map[string]any{
				{
					"id":                     3001,
					"body":                   "inline comment",
					"pull_request_review_id": 2001,
					"created_at":             "2026-01-03T02:00:00Z",
					"updated_at":             "2026-01-03T02:00:00Z",
					"html_url":               "https://github.com/octo/repo/pull/2#discussion_r3001",
					"user":                   map[string]any{"login": "reviewer"},
				},
				{
					"id":                     3002,
					"body":                   "orphan inline comment",
					"pull_request_review_id": 9999,
					"created_at":             "2026-01-03T03:00:00Z",
					"updated_at":             "2026-01-03T03:00:00Z",
					"html_url":               "https://github.com/octo/repo/pull/2#discussion_r3002",
					"user":                   map[string]any{"login": "orphan-user"},
				},
			}), nil
		default:
			return notFoundResponse(r.URL.Path), nil
		}
	})

	fetcher, err := NewFetcher(Config{
		HTTPClient:  clientHTTP,
		RESTBaseURL: "https://api.test/",
		GraphQLURL:  "https://api.test/graphql",
	})
	if err != nil {
		t.Fatalf("NewFetcher error = %v, want nil", err)
	}

	got, err := fetcher.Fetch(context.Background(), ResourceRef{
		Owner:  "octo",
		Repo:   "repo",
		Number: 2,
		Type:   ResourcePullRequest,
		URL:    "https://github.com/octo/repo/pull/2",
	}, FetchOptions{IncludeComments: true})
	if err != nil {
		t.Fatalf("Fetch error = %v, want nil", err)
	}

	if got.Meta.Type != ResourcePullRequest {
		t.Fatalf("Meta.Type = %q, want %q", got.Meta.Type, ResourcePullRequest)
	}
	if !got.Meta.Merged {
		t.Fatal("Meta.Merged = false, want true")
	}
	if len(got.Reviews) != 1 {
		t.Fatalf("reviews len = %d, want 1", len(got.Reviews))
	}
	if len(got.Reviews[0].Comments) != 1 {
		t.Fatalf("review comments len = %d, want 1", len(got.Reviews[0].Comments))
	}
	if got.Reactions.Total != 4 {
		t.Fatalf("top reactions total = %d, want 4", got.Reactions.Total)
	}
	if len(got.Thread) != 2 {
		t.Fatalf("thread len = %d, want 2 (issue comment + orphan review comment)", len(got.Thread))
	}
	if got.Thread[0].Body != "top-level PR conversation comment" {
		t.Fatalf("first thread body = %q, want %q", got.Thread[0].Body, "top-level PR conversation comment")
	}
	if got.Thread[1].Body != "orphan inline comment" {
		t.Fatalf("second thread body = %q, want %q", got.Thread[1].Body, "orphan inline comment")
	}
}
