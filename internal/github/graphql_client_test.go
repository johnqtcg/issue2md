package github

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

func TestGraphQLQueryRequestShapeAndAuth(t *testing.T) {
	t.Parallel()

	clientHTTP := newTestHTTPClient(func(r *http.Request) (*http.Response, error) {
		if r.Method != http.MethodPost {
			t.Fatalf("method = %s, want POST", r.Method)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer gql-token" {
			t.Fatalf("Authorization header = %q, want %q", got, "Bearer gql-token")
		}

		var req map[string]any
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if req["query"] != "query Test($x:Int!){value}" {
			t.Fatalf("query = %v, want exact query", req["query"])
		}
		variables := req["variables"].(map[string]any)
		if variables["x"] != float64(7) {
			t.Fatalf("variables.x = %v, want 7", variables["x"])
		}

		return mustJSONResponse(t, http.StatusOK, map[string]any{
			"data": map[string]any{
				"value": "ok",
			},
		}), nil
	})

	client := newGraphQLClient(Config{
		Token:      "gql-token",
		HTTPClient: clientHTTP,
		GraphQLURL: "https://api.test/graphql",
	})

	var out struct {
		Value string `json:"value"`
	}
	err := client.Query(context.Background(), "query Test($x:Int!){value}", map[string]any{"x": 7}, &out)
	if err != nil {
		t.Fatalf("Query error = %v, want nil", err)
	}
	if out.Value != "ok" {
		t.Fatalf("out.Value = %q, want ok", out.Value)
	}
}

func TestGraphQLQueryPaginatedPassesCursor(t *testing.T) {
	t.Parallel()

	call := 0
	clientHTTP := newTestHTTPClient(func(r *http.Request) (*http.Response, error) {
		call++
		var req struct {
			Variables map[string]any `json:"variables"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request: %v", err)
		}

		if call == 1 {
			if _, ok := req.Variables["after"]; ok {
				t.Fatalf("first call unexpectedly includes after: %v", req.Variables["after"])
			}
			return mustJSONResponse(t, http.StatusOK, map[string]any{
				"data": map[string]any{
					"pageInfo": map[string]any{
						"hasNextPage": true,
						"endCursor":   "c1",
					},
				},
			}), nil
		}
		if call == 2 {
			if got := req.Variables["after"]; got != "c1" {
				t.Fatalf("second call after = %v, want c1", got)
			}
			return mustJSONResponse(t, http.StatusOK, map[string]any{
				"data": map[string]any{
					"pageInfo": map[string]any{
						"hasNextPage": false,
						"endCursor":   "",
					},
				},
			}), nil
		}

		t.Fatalf("unexpected call count: %d", call)
		return nil, nil
	})

	client := newGraphQLClient(Config{
		HTTPClient: clientHTTP,
		GraphQLURL: "https://api.test/graphql",
	})

	err := client.QueryPaginated(context.Background(), "query P { pageInfo { hasNextPage endCursor } }", map[string]any{}, func(page json.RawMessage) (bool, string, error) {
		var out struct {
			PageInfo struct {
				HasNextPage bool   `json:"hasNextPage"`
				EndCursor   string `json:"endCursor"`
			} `json:"pageInfo"`
		}
		if err := json.Unmarshal(page, &out); err != nil {
			return false, "", err
		}
		return out.PageInfo.HasNextPage, out.PageInfo.EndCursor, nil
	})
	if err != nil {
		t.Fatalf("QueryPaginated error = %v, want nil", err)
	}
	if call != 2 {
		t.Fatalf("call count = %d, want 2", call)
	}
}

func TestGraphQLQueryDecodeError(t *testing.T) {
	t.Parallel()

	clientHTTP := newTestHTTPClient(func(r *http.Request) (*http.Response, error) {
		return textHTTPResponse(http.StatusOK, "{not-json"), nil
	})

	client := newGraphQLClient(Config{
		HTTPClient: clientHTTP,
		GraphQLURL: "https://api.test/graphql",
	})

	var out map[string]any
	err := client.Query(context.Background(), "query{}", nil, &out)
	if err == nil {
		t.Fatal("Query error = nil, want error")
	}
}

func TestGraphQLQueryPaginatedCursorStalled(t *testing.T) {
	t.Parallel()

	call := 0
	clientHTTP := newTestHTTPClient(func(r *http.Request) (*http.Response, error) {
		call++
		return mustJSONResponse(t, http.StatusOK, map[string]any{
			"data": map[string]any{
				"pageInfo": map[string]any{
					"hasNextPage": true,
					"endCursor":   "same-cursor",
				},
			},
		}), nil
	})

	client := newGraphQLClient(Config{
		HTTPClient: clientHTTP,
		GraphQLURL: "https://api.test/graphql",
	})

	err := client.QueryPaginated(context.Background(), "query P { pageInfo { hasNextPage endCursor } }", nil, func(page json.RawMessage) (bool, string, error) {
		var out struct {
			PageInfo struct {
				HasNextPage bool   `json:"hasNextPage"`
				EndCursor   string `json:"endCursor"`
			} `json:"pageInfo"`
		}
		if err := json.Unmarshal(page, &out); err != nil {
			return false, "", err
		}
		return out.PageInfo.HasNextPage, out.PageInfo.EndCursor, nil
	})
	if err == nil {
		t.Fatal("QueryPaginated error = nil, want error")
	}
	if !strings.Contains(err.Error(), "cursor stalled") {
		t.Fatalf("error = %v, want contains %q", err, "cursor stalled")
	}
	if call != 2 {
		t.Fatalf("call count = %d, want 2", call)
	}
}
