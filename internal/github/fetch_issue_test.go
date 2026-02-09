package github

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

func TestFetchIssue(t *testing.T) {
	t.Parallel()

	graphqlCalls := 0
	clientHTTP := newTestHTTPClient(func(r *http.Request) (*http.Response, error) {
		switch r.URL.Path {
		case "/repos/octo/repo/issues/1":
			return mustJSONResponse(t, http.StatusOK, map[string]any{
				"number":     1,
				"title":      "Issue title",
				"state":      "open",
				"body":       "Issue body",
				"html_url":   "https://github.com/octo/repo/issues/1",
				"created_at": "2026-01-01T00:00:00Z",
				"updated_at": "2026-01-02T00:00:00Z",
				"user":       map[string]any{"login": "alice"},
				"labels":     []map[string]any{{"name": "bug"}},
				"reactions": map[string]any{
					"+1":          2,
					"-1":          1,
					"heart":       1,
					"total_count": 4,
				},
			}), nil
		case "/repos/octo/repo/issues/1/comments":
			return mustJSONResponse(t, http.StatusOK, []map[string]any{
				{
					"id":         1001,
					"body":       "comment one",
					"html_url":   "https://github.com/octo/repo/issues/1#issuecomment-1001",
					"created_at": "2026-01-03T00:00:00Z",
					"updated_at": "2026-01-03T00:00:00Z",
					"user":       map[string]any{"login": "bob"},
					"reactions": map[string]any{
						"heart":       2,
						"total_count": 2,
					},
				},
			}), nil
		case "/graphql":
			graphqlCalls++
			var req struct {
				Variables map[string]any `json:"variables"`
			}
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				t.Fatalf("decode graphql request: %v", err)
			}
			if graphqlCalls == 1 {
				if _, ok := req.Variables["after"]; ok {
					t.Fatalf("first issue timeline page must not include cursor")
				}
				return mustJSONResponse(t, http.StatusOK, map[string]any{
					"data": map[string]any{
						"repository": map[string]any{
							"issue": map[string]any{
								"timelineItems": map[string]any{
									"nodes": []map[string]any{
										{
											"__typename": "ClosedEvent",
											"createdAt":  "2026-01-04T00:00:00Z",
											"actor":      map[string]any{"login": "carol"},
										},
									},
									"pageInfo": map[string]any{
										"hasNextPage": true,
										"endCursor":   "cursor-1",
									},
								},
							},
						},
					},
				}), nil
			}
			if req.Variables["after"] != "cursor-1" {
				t.Fatalf("second issue timeline page after = %v, want cursor-1", req.Variables["after"])
			}
			return mustJSONResponse(t, http.StatusOK, map[string]any{
				"data": map[string]any{
					"repository": map[string]any{
						"issue": map[string]any{
							"timelineItems": map[string]any{
								"nodes": []map[string]any{
									{
										"__typename": "LabeledEvent",
										"createdAt":  "2026-01-05T00:00:00Z",
										"actor":      map[string]any{"login": "dora"},
										"label":      map[string]any{"name": "bug"},
									},
								},
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

	got, err := fetcher.Fetch(context.Background(), ResourceRef{
		Owner:  "octo",
		Repo:   "repo",
		Number: 1,
		Type:   ResourceIssue,
		URL:    "https://github.com/octo/repo/issues/1",
	}, FetchOptions{IncludeComments: true})
	if err != nil {
		t.Fatalf("Fetch error = %v, want nil", err)
	}

	if got.Meta.Type != ResourceIssue {
		t.Fatalf("Meta.Type = %q, want %q", got.Meta.Type, ResourceIssue)
	}
	if got.Reactions.Total != 4 {
		t.Fatalf("top reactions total = %d, want 4", got.Reactions.Total)
	}
	if len(got.Thread) != 1 {
		t.Fatalf("thread len = %d, want 1", len(got.Thread))
	}
	if got.Thread[0].Reactions.Heart != 2 {
		t.Fatalf("comment heart reactions = %d, want 2", got.Thread[0].Reactions.Heart)
	}
	if !hasTimelineEvent(got.Timeline, "opened") || !hasTimelineEvent(got.Timeline, "closed") {
		t.Fatalf("timeline missing required events: %#v", got.Timeline)
	}
	if !hasTimelineEvent(got.Timeline, "labeled") {
		t.Fatalf("timeline missing labeled event from paginated page: %#v", got.Timeline)
	}
	if graphqlCalls != 2 {
		t.Fatalf("graphql calls = %d, want 2", graphqlCalls)
	}
}

func TestFetchIssueDoesNotDuplicateOpenedTimelineEvent(t *testing.T) {
	t.Parallel()

	clientHTTP := newTestHTTPClient(func(r *http.Request) (*http.Response, error) {
		switch r.URL.Path {
		case "/repos/octo/repo/issues/1":
			return mustJSONResponse(t, http.StatusOK, map[string]any{
				"number":     1,
				"title":      "Issue title",
				"state":      "open",
				"body":       "Issue body",
				"html_url":   "https://github.com/octo/repo/issues/1",
				"created_at": "2026-01-01T00:00:00Z",
				"updated_at": "2026-01-01T00:00:00Z",
				"user":       map[string]any{"login": "alice"},
			}), nil
		case "/graphql":
			return mustJSONResponse(t, http.StatusOK, map[string]any{
				"data": map[string]any{
					"repository": map[string]any{
						"issue": map[string]any{
							"timelineItems": map[string]any{
								"nodes": []map[string]any{
									{
										"__typename": "OpenedEvent",
										"createdAt":  "2026-01-01T00:00:00Z",
										"actor":      map[string]any{"login": "alice"},
									},
								},
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

	got, err := fetcher.Fetch(context.Background(), ResourceRef{
		Owner:  "octo",
		Repo:   "repo",
		Number: 1,
		Type:   ResourceIssue,
		URL:    "https://github.com/octo/repo/issues/1",
	}, FetchOptions{IncludeComments: false})
	if err != nil {
		t.Fatalf("Fetch error = %v, want nil", err)
	}

	if count := countTimelineEvent(got.Timeline, "opened"); count != 1 {
		t.Fatalf("opened event count = %d, want 1; timeline=%#v", count, got.Timeline)
	}
}

func TestFetchIssueTimelineUsesUnionFragments(t *testing.T) {
	t.Parallel()

	clientHTTP := newTestHTTPClient(func(r *http.Request) (*http.Response, error) {
		switch r.URL.Path {
		case "/repos/octo/repo/issues/1":
			return mustJSONResponse(t, http.StatusOK, map[string]any{
				"number":     1,
				"title":      "Issue title",
				"state":      "open",
				"body":       "Issue body",
				"html_url":   "https://github.com/octo/repo/issues/1",
				"created_at": "2026-01-01T00:00:00Z",
				"updated_at": "2026-01-01T00:00:00Z",
				"user":       map[string]any{"login": "alice"},
			}), nil
		case "/graphql":
			var req struct {
				Query string `json:"query"`
			}
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				t.Fatalf("decode graphql request: %v", err)
			}

			if strings.Contains(req.Query, "nodes {\n          __typename\n          createdAt") ||
				strings.Contains(req.Query, "nodes {\n          __typename\n          actor { login }") ||
				strings.Contains(req.Query, "assignee { login }") ||
				strings.Contains(req.Query, "milestone { title }") {
				return mustJSONResponse(t, http.StatusOK, map[string]any{
					"errors": []map[string]any{
						{
							"message": "Selections can't be made directly on unions (see selections on Assignee)",
						},
					},
				}), nil
			}

			return mustJSONResponse(t, http.StatusOK, map[string]any{
				"data": map[string]any{
					"repository": map[string]any{
						"issue": map[string]any{
							"timelineItems": map[string]any{
								"nodes": []map[string]any{
									{
										"__typename": "ClosedEvent",
										"createdAt":  "2026-01-02T00:00:00Z",
										"actor":      map[string]any{"login": "bob"},
									},
								},
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

	got, err := fetcher.Fetch(context.Background(), ResourceRef{
		Owner:  "octo",
		Repo:   "repo",
		Number: 1,
		Type:   ResourceIssue,
		URL:    "https://github.com/octo/repo/issues/1",
	}, FetchOptions{IncludeComments: false})
	if err != nil {
		t.Fatalf("Fetch error = %v, want nil", err)
	}

	if !hasTimelineEvent(got.Timeline, "closed") {
		t.Fatalf("timeline missing closed event: %#v", got.Timeline)
	}
}

func hasTimelineEvent(events []TimelineEvent, want string) bool {
	for _, event := range events {
		if event.EventType == want {
			return true
		}
	}
	return false
}

func countTimelineEvent(events []TimelineEvent, want string) int {
	count := 0
	for _, event := range events {
		if event.EventType == want {
			count++
		}
	}
	return count
}
