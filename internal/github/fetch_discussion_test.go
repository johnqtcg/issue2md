package github

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"testing"
)

func TestFetchDiscussion(t *testing.T) {
	t.Parallel()

	discussionPageCalls := 0
	repliesCalls := 0
	clientHTTP := newTestHTTPClient(func(r *http.Request) (*http.Response, error) {
		if r.URL.Path != "/graphql" {
			return notFoundResponse(r.URL.Path), nil
		}

		var req struct {
			Query     string         `json:"query"`
			Variables map[string]any `json:"variables"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode graphql request: %v", err)
		}

		if req.Variables["commentID"] != nil {
			repliesCalls++
			if repliesCalls == 1 {
				if req.Variables["after"] != "r-cursor-1" {
					t.Fatalf("first reply page after = %v, want r-cursor-1", req.Variables["after"])
				}
				return mustJSONResponse(t, http.StatusOK, map[string]any{
					"data": map[string]any{
						"node": map[string]any{
							"replies": map[string]any{
								"nodes": []map[string]any{
									{
										"id":        "r2",
										"body":      "reply-2",
										"createdAt": "2026-01-03T02:00:00Z",
										"updatedAt": "2026-01-03T02:00:00Z",
										"url":       "https://github.com/octo/repo/discussions/9#discussioncomment-22",
										"author":    map[string]any{"login": "eve"},
										"reactions": map[string]any{"total": 0},
									},
								},
								"pageInfo": map[string]any{
									"hasNextPage": false,
									"endCursor":   "",
								},
							},
						},
					},
				}), nil
			}
			t.Fatalf("unexpected discussion replies call: %d", repliesCalls)
		}

		discussionPageCalls++
		if discussionPageCalls == 1 {
			if _, ok := req.Variables["after"]; ok {
				t.Fatalf("first page must not include cursor, got after=%v", req.Variables["after"])
			}
			return mustJSONResponse(t, http.StatusOK, map[string]any{
				"data": map[string]any{
					"repository": map[string]any{
						"discussion": map[string]any{
							"number":     9,
							"title":      "Discussion title",
							"body":       "Discussion body",
							"url":        "https://github.com/octo/repo/discussions/9",
							"createdAt":  "2026-01-01T00:00:00Z",
							"updatedAt":  "2026-01-02T00:00:00Z",
							"closed":     false,
							"author":     map[string]any{"login": "alice"},
							"category":   map[string]any{"name": "Q&A"},
							"isAnswered": true,
							"answer": map[string]any{
								"id":     "answer-1",
								"author": map[string]any{"login": "bob"},
							},
							"reactions": map[string]any{
								"plusOne": 2,
								"heart":   1,
								"total":   3,
							},
							"comments": map[string]any{
								"nodes": []map[string]any{
									{
										"id":        "c1",
										"body":      "first",
										"createdAt": "2026-01-03T00:00:00Z",
										"updatedAt": "2026-01-03T00:00:00Z",
										"url":       "https://github.com/octo/repo/discussions/9#discussioncomment-1",
										"author":    map[string]any{"login": "carol"},
										"reactions": map[string]any{"heart": 1, "total": 1},
										"replies": map[string]any{
											"nodes": []map[string]any{
												{
													"id":        "r1",
													"body":      "reply-1",
													"createdAt": "2026-01-03T01:00:00Z",
													"updatedAt": "2026-01-03T01:00:00Z",
													"url":       "https://github.com/octo/repo/discussions/9#discussioncomment-2",
													"author":    map[string]any{"login": "dave"},
													"reactions": map[string]any{"total": 0},
												},
											},
											"pageInfo": map[string]any{
												"hasNextPage": true,
												"endCursor":   "r-cursor-1",
											},
										},
									},
								},
								"pageInfo": map[string]any{
									"hasNextPage": true,
									"endCursor":   "c1",
								},
							},
						},
					},
				},
			}), nil
		}

		if got := req.Variables["after"]; got != "c1" {
			t.Fatalf("second page after = %v, want c1", got)
		}
		return mustJSONResponse(t, http.StatusOK, map[string]any{
			"data": map[string]any{
				"repository": map[string]any{
					"discussion": map[string]any{
						"comments": map[string]any{
							"nodes": []map[string]any{
								{
									"id":        "c2",
									"body":      "second",
									"createdAt": "2026-01-04T00:00:00Z",
									"updatedAt": "2026-01-04T00:00:00Z",
									"url":       "https://github.com/octo/repo/discussions/9#discussioncomment-3",
									"author":    map[string]any{"login": "erin"},
									"reactions": map[string]any{"total": 0},
									"replies": map[string]any{
										"nodes": []map[string]any{},
										"pageInfo": map[string]any{
											"hasNextPage": false,
											"endCursor":   "",
										},
									},
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
	})

	fetcher, err := NewFetcher(Config{
		HTTPClient: clientHTTP,
		GraphQLURL: "https://api.test/graphql",
	})
	if err != nil {
		t.Fatalf("NewFetcher error = %v, want nil", err)
	}

	got, err := fetcher.Fetch(context.Background(), ResourceRef{
		Owner:  "octo",
		Repo:   "repo",
		Number: 9,
		Type:   ResourceDiscussion,
		URL:    "https://github.com/octo/repo/discussions/9",
	}, FetchOptions{IncludeComments: true})
	if err != nil {
		t.Fatalf("Fetch error = %v, want nil", err)
	}

	if got.Meta.Type != ResourceDiscussion {
		t.Fatalf("Meta.Type = %q, want %q", got.Meta.Type, ResourceDiscussion)
	}
	if !got.Meta.IsAnswered {
		t.Fatal("Meta.IsAnswered = false, want true")
	}
	if got.Meta.AcceptedAnswerAuthor != "bob" {
		t.Fatalf("AcceptedAnswerAuthor = %q, want bob", got.Meta.AcceptedAnswerAuthor)
	}
	if got.Meta.AcceptedAnswerID != "answer-1" {
		t.Fatalf("AcceptedAnswerID = %q, want answer-1", got.Meta.AcceptedAnswerID)
	}
	if got.Reactions.Total != 3 {
		t.Fatalf("top reaction total = %d, want 3", got.Reactions.Total)
	}
	if len(got.Thread) != 2 {
		t.Fatalf("thread len = %d, want 2", len(got.Thread))
	}
	if len(got.Thread[0].Replies) != 2 {
		t.Fatalf("first thread reply len = %d, want 2", len(got.Thread[0].Replies))
	}
	if repliesCalls != 1 {
		t.Fatalf("replies calls = %d, want 1", repliesCalls)
	}
}

func TestFetchDiscussionNotFound(t *testing.T) {
	t.Parallel()

	clientHTTP := newTestHTTPClient(func(r *http.Request) (*http.Response, error) {
		return mustJSONResponse(t, http.StatusOK, map[string]any{
			"data": map[string]any{
				"repository": map[string]any{
					"discussion": nil,
				},
			},
		}), nil
	})

	fetcher, err := NewFetcher(Config{
		HTTPClient: clientHTTP,
		GraphQLURL: "https://api.test/graphql",
	})
	if err != nil {
		t.Fatalf("NewFetcher error = %v, want nil", err)
	}

	_, err = fetcher.Fetch(context.Background(), ResourceRef{
		Owner:  "octo",
		Repo:   "repo",
		Number: 9,
		Type:   ResourceDiscussion,
		URL:    "https://github.com/octo/repo/discussions/9",
	}, FetchOptions{IncludeComments: true})
	if err == nil {
		t.Fatal("Fetch error = nil, want error")
	}
	if !errors.Is(err, ErrResourceNotFound) {
		t.Fatalf("Fetch error = %v, want ErrResourceNotFound", err)
	}
}

func TestFetchDiscussionSkipsCommentPaginationWhenCommentsDisabled(t *testing.T) {
	t.Parallel()

	discussionPageCalls := 0
	clientHTTP := newTestHTTPClient(func(r *http.Request) (*http.Response, error) {
		if r.URL.Path != "/graphql" {
			return notFoundResponse(r.URL.Path), nil
		}

		var req struct {
			Query     string         `json:"query"`
			Variables map[string]any `json:"variables"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode graphql request: %v", err)
		}

		discussionPageCalls++
		if strings.Contains(req.Query, "comments(first:50") {
			t.Fatalf("query should not include comments section when include-comments=false")
		}

		return mustJSONResponse(t, http.StatusOK, map[string]any{
			"data": map[string]any{
				"repository": map[string]any{
					"discussion": map[string]any{
						"number":     9,
						"title":      "Discussion title",
						"body":       "Discussion body",
						"url":        "https://github.com/octo/repo/discussions/9",
						"createdAt":  "2026-01-01T00:00:00Z",
						"updatedAt":  "2026-01-02T00:00:00Z",
						"closed":     false,
						"author":     map[string]any{"login": "alice"},
						"category":   map[string]any{"name": "Q&A"},
						"isAnswered": false,
						"reactions": map[string]any{
							"plusOne": 1,
							"heart":   0,
							"total":   1,
						},
					},
				},
			},
		}), nil
	})

	fetcher, err := NewFetcher(Config{
		HTTPClient: clientHTTP,
		GraphQLURL: "https://api.test/graphql",
	})
	if err != nil {
		t.Fatalf("NewFetcher error = %v, want nil", err)
	}

	got, err := fetcher.Fetch(context.Background(), ResourceRef{
		Owner:  "octo",
		Repo:   "repo",
		Number: 9,
		Type:   ResourceDiscussion,
		URL:    "https://github.com/octo/repo/discussions/9",
	}, FetchOptions{IncludeComments: false})
	if err != nil {
		t.Fatalf("Fetch error = %v, want nil", err)
	}
	if len(got.Thread) != 0 {
		t.Fatalf("thread len = %d, want 0", len(got.Thread))
	}
	if discussionPageCalls != 1 {
		t.Fatalf("graphql discussion page calls = %d, want 1", discussionPageCalls)
	}
}
