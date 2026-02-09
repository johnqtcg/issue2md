package github

import (
	"context"
	"errors"
	"net/http"
	"testing"
)

func TestFetcherDispatchByResourceType(t *testing.T) {
	t.Parallel()

	clientHTTP := newTestHTTPClient(func(r *http.Request) (*http.Response, error) {
		switch r.URL.Path {
		case "/repos/octo/repo/issues/1":
			return mustJSONResponse(t, http.StatusOK, map[string]any{
				"number":     1,
				"title":      "Issue",
				"state":      "open",
				"body":       "Issue body",
				"html_url":   "https://github.com/octo/repo/issues/1",
				"created_at": "2026-01-01T00:00:00Z",
				"updated_at": "2026-01-01T00:00:00Z",
				"user":       map[string]any{"login": "alice"},
			}), nil
		case "/repos/octo/repo/issues/2":
			return mustJSONResponse(t, http.StatusOK, map[string]any{
				"number":     2,
				"title":      "PR as issue",
				"state":      "open",
				"body":       "PR issue envelope",
				"html_url":   "https://github.com/octo/repo/issues/2",
				"created_at": "2026-01-01T00:00:00Z",
				"updated_at": "2026-01-01T00:00:00Z",
				"user":       map[string]any{"login": "alice"},
				"reactions":  map[string]any{"total_count": 0},
			}), nil
		case "/repos/octo/repo/issues/1/comments":
			return mustJSONResponse(t, http.StatusOK, []map[string]any{}), nil
		case "/repos/octo/repo/pulls/2":
			return mustJSONResponse(t, http.StatusOK, map[string]any{
				"number":     2,
				"title":      "PR",
				"state":      "open",
				"body":       "PR body",
				"html_url":   "https://github.com/octo/repo/pull/2",
				"created_at": "2026-01-01T00:00:00Z",
				"updated_at": "2026-01-01T00:00:00Z",
				"user":       map[string]any{"login": "alice"},
			}), nil
		case "/repos/octo/repo/pulls/2/reviews":
			return mustJSONResponse(t, http.StatusOK, []map[string]any{}), nil
		case "/repos/octo/repo/pulls/2/comments":
			return mustJSONResponse(t, http.StatusOK, []map[string]any{}), nil
		case "/graphql":
			return mustJSONResponse(t, http.StatusOK, map[string]any{
				"data": map[string]any{
					"repository": map[string]any{
						"issue": map[string]any{
							"timelineItems": map[string]any{"nodes": []map[string]any{}},
						},
						"pullRequest": map[string]any{
							"reviewThreads": map[string]any{"nodes": []map[string]any{}},
						},
						"discussion": map[string]any{
							"number":     9,
							"title":      "Discussion",
							"body":       "Discussion body",
							"url":        "https://github.com/octo/repo/discussions/9",
							"createdAt":  "2026-01-01T00:00:00Z",
							"updatedAt":  "2026-01-01T00:00:00Z",
							"closed":     false,
							"author":     map[string]any{"login": "alice"},
							"category":   map[string]any{"name": "Q&A"},
							"isAnswered": false,
							"reactions":  map[string]any{"total": 0},
							"comments": map[string]any{
								"nodes": []map[string]any{},
								"pageInfo": map[string]any{
									"hasNextPage": false,
									"endCursor":   "",
								},
							},
						},
					},
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

	tcs := []struct {
		name string
		ref  ResourceRef
		want ResourceType
	}{
		{name: "issue", ref: ResourceRef{Owner: "octo", Repo: "repo", Number: 1, Type: ResourceIssue}, want: ResourceIssue},
		{name: "pull request", ref: ResourceRef{Owner: "octo", Repo: "repo", Number: 2, Type: ResourcePullRequest}, want: ResourcePullRequest},
		{name: "discussion", ref: ResourceRef{Owner: "octo", Repo: "repo", Number: 9, Type: ResourceDiscussion}, want: ResourceDiscussion},
	}

	for _, tc := range tcs {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			got, err := fetcher.Fetch(context.Background(), tc.ref, FetchOptions{IncludeComments: false})
			if err != nil {
				t.Fatalf("Fetch error = %v, want nil", err)
			}
			if got.Meta.Type != tc.want {
				t.Fatalf("Meta.Type = %q, want %q", got.Meta.Type, tc.want)
			}
		})
	}
}

func TestFetcherIncludeCommentsOption(t *testing.T) {
	t.Parallel()

	clientHTTP := newTestHTTPClient(func(r *http.Request) (*http.Response, error) {
		switch r.URL.Path {
		case "/repos/octo/repo/issues/1":
			return mustJSONResponse(t, http.StatusOK, map[string]any{
				"number":     1,
				"title":      "Issue",
				"state":      "open",
				"body":       "Issue body",
				"html_url":   "https://github.com/octo/repo/issues/1",
				"created_at": "2026-01-01T00:00:00Z",
				"updated_at": "2026-01-01T00:00:00Z",
				"user":       map[string]any{"login": "alice"},
			}), nil
		case "/repos/octo/repo/issues/1/comments":
			return mustJSONResponse(t, http.StatusOK, []map[string]any{
				{
					"id":         1,
					"body":       "comment",
					"created_at": "2026-01-01T00:00:00Z",
					"updated_at": "2026-01-01T00:00:00Z",
					"user":       map[string]any{"login": "bob"},
				},
			}), nil
		case "/graphql":
			return mustJSONResponse(t, http.StatusOK, map[string]any{
				"data": map[string]any{
					"repository": map[string]any{
						"issue": map[string]any{
							"timelineItems": map[string]any{"nodes": []map[string]any{}},
						},
					},
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

	ref := ResourceRef{Owner: "octo", Repo: "repo", Number: 1, Type: ResourceIssue}

	withComments, err := fetcher.Fetch(context.Background(), ref, FetchOptions{IncludeComments: true})
	if err != nil {
		t.Fatalf("Fetch with comments error = %v, want nil", err)
	}
	withoutComments, err := fetcher.Fetch(context.Background(), ref, FetchOptions{IncludeComments: false})
	if err != nil {
		t.Fatalf("Fetch without comments error = %v, want nil", err)
	}

	if len(withComments.Thread) != 1 {
		t.Fatalf("with comments thread len = %d, want 1", len(withComments.Thread))
	}
	if len(withoutComments.Thread) != 0 {
		t.Fatalf("without comments thread len = %d, want 0", len(withoutComments.Thread))
	}
}

func TestFetcherUnsupportedType(t *testing.T) {
	t.Parallel()

	fetcher, err := NewFetcher(Config{})
	if err != nil {
		t.Fatalf("NewFetcher error = %v, want nil", err)
	}

	_, err = fetcher.Fetch(context.Background(), ResourceRef{
		Type: ResourceType("unknown"),
	}, FetchOptions{})
	if !errors.Is(err, ErrUnsupportedResourceType) {
		t.Fatalf("Fetch error = %v, want ErrUnsupportedResourceType", err)
	}
}
