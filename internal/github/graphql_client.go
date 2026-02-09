package github

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const defaultGraphQLURL = "https://api.github.com/graphql"
const maxGraphQLPages = 1000

type graphQLClient struct {
	httpClient *http.Client
	endpoint   string
	token      string
}

type graphQLRequest struct {
	Query     string         `json:"query"`
	Variables map[string]any `json:"variables,omitempty"`
}

type graphQLErrorMessage struct {
	Message string `json:"message"`
}

type graphQLResponse struct {
	Data   json.RawMessage       `json:"data"`
	Errors []graphQLErrorMessage `json:"errors"`
}

func newGraphQLClient(cfg Config) *graphQLClient {
	httpClient := cfg.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 30 * time.Second}
	}

	endpoint := cfg.GraphQLURL
	if endpoint == "" {
		endpoint = defaultGraphQLURL
	}

	return &graphQLClient{
		httpClient: httpClient,
		endpoint:   endpoint,
		token:      cfg.Token,
	}
}

func (c *graphQLClient) Query(ctx context.Context, query string, variables map[string]any, out any) error {
	body, err := c.queryRaw(ctx, query, variables)
	if err != nil {
		return err
	}
	if out == nil {
		return nil
	}
	if err := json.Unmarshal(body, out); err != nil {
		return fmt.Errorf("decode graphql data: %w", err)
	}
	return nil
}

func (c *graphQLClient) QueryPaginated(ctx context.Context, query string, variables map[string]any, pageHandler func(page json.RawMessage) (hasNext bool, endCursor string, err error)) error {
	currentVars := copyVariables(variables)
	previousCursor := ""
	for pageIndex := 0; ; pageIndex++ {
		if pageIndex >= maxGraphQLPages {
			return fmt.Errorf("graphql pagination exceeded max page limit %d", maxGraphQLPages)
		}

		page, err := c.queryRaw(ctx, query, currentVars)
		if err != nil {
			return err
		}
		hasNext, endCursor, err := pageHandler(page)
		if err != nil {
			return fmt.Errorf("handle paginated graphql page: %w", err)
		}
		if !hasNext {
			return nil
		}
		if endCursor == "" {
			return fmt.Errorf("graphql pagination returned empty cursor while hasNextPage=true")
		}
		if endCursor == previousCursor {
			return fmt.Errorf("graphql pagination cursor stalled at %q", endCursor)
		}
		currentVars["after"] = endCursor
		previousCursor = endCursor
	}
}

func (c *graphQLClient) queryRaw(ctx context.Context, query string, variables map[string]any) (json.RawMessage, error) {
	payload := graphQLRequest{
		Query:     query,
		Variables: variables,
	}
	requestBody, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal graphql request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint, bytes.NewReader(requestBody))
	if err != nil {
		return nil, fmt.Errorf("create graphql request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("execute graphql request: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 16*1024))
		return nil, fmt.Errorf("graphql status error: %w", &statusError{
			StatusCode: resp.StatusCode,
			Err:        fmt.Errorf("%s", strings.TrimSpace(string(body))),
		})
	}

	var envelope graphQLResponse
	if err := json.NewDecoder(resp.Body).Decode(&envelope); err != nil {
		return nil, fmt.Errorf("decode graphql response: %w", err)
	}
	if len(envelope.Errors) > 0 {
		return nil, fmt.Errorf("graphql returned errors: %s", envelope.Errors[0].Message)
	}

	return envelope.Data, nil
}

func copyVariables(in map[string]any) map[string]any {
	if len(in) == 0 {
		return map[string]any{}
	}
	out := make(map[string]any, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}
