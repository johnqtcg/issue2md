package converter

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"testing"
)

func TestOpenAISummarizerSuccess(t *testing.T) {
	t.Parallel()

	clientHTTP := &http.Client{
		Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			if r.Method != http.MethodPost {
				t.Fatalf("method = %s, want POST", r.Method)
			}
			if got := r.Header.Get("Authorization"); got != "Bearer test-key" {
				t.Fatalf("Authorization = %q, want Bearer test-key", got)
			}
			if r.URL.String() != "https://api.test/v1/responses" {
				t.Fatalf("url = %s, want https://api.test/v1/responses", r.URL.String())
			}

			var payload map[string]any
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				t.Fatalf("decode request: %v", err)
			}
			if payload["model"] != "gpt-test" {
				t.Fatalf("model = %v, want gpt-test", payload["model"])
			}
			input, _ := payload["input"].(string)
			if !strings.Contains(input, "Language: zh") {
				t.Fatalf("prompt should include language override, got: %s", input)
			}

			return mustJSONResponse(http.StatusOK, map[string]any{
				"output": []map[string]any{
					{
						"type": "message",
						"content": []map[string]any{
							{
								"type": "output_text",
								"text": `{"summary":"brief","key_decisions":["d1"],"action_items":["a1"],"language":"zh"}`,
							},
						},
					},
				},
			}), nil
		}),
	}

	s := NewOpenAISummarizer(OpenAISummarizerConfig{
		APIKey:     "test-key",
		BaseURL:    "https://api.test",
		Model:      "gpt-test",
		HTTPClient: clientHTTP,
	})

	got, err := s.Summarize(context.Background(), sampleIssueData(), "zh")
	if err != nil {
		t.Fatalf("Summarize error = %v, want nil", err)
	}
	if got.Status != "ok" {
		t.Fatalf("Status = %q, want ok", got.Status)
	}
	if got.Summary != "brief" {
		t.Fatalf("Summary = %q, want brief", got.Summary)
	}
	if got.Language != "zh" {
		t.Fatalf("Language = %q, want zh", got.Language)
	}
}

func TestResolveSummaryLanguage(t *testing.T) {
	t.Parallel()

	tcs := []struct {
		name     string
		override string
		data     func() string
		want     string
	}{
		{
			name:     "lang override wins",
			override: "zh",
			data: func() string {
				return "English content."
			},
			want: "zh",
		},
		{
			name:     "auto detect chinese",
			override: "",
			data: func() string {
				return "这是中文讨论内容。"
			},
			want: "zh",
		},
		{
			name:     "auto detect english fallback",
			override: "",
			data: func() string {
				return "This is English text."
			},
			want: "en",
		},
	}

	for _, tc := range tcs {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			d := sampleIssueData()
			d.Description = tc.data()
			got := resolveSummaryLanguage(tc.override, d)
			if got != tc.want {
				t.Fatalf("resolveSummaryLanguage = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestRendererSummaryDegradePath(t *testing.T) {
	t.Parallel()

	r := NewRenderer(&stubSummarizer{err: errors.New("provider timeout")})
	out, err := r.Render(context.Background(), sampleIssueData(), RenderOptions{
		IncludeComments: true,
		IncludeSummary:  true,
	})
	if err != nil {
		t.Fatalf("Render error = %v, want nil", err)
	}
	content := string(out)
	if strings.Contains(content, "## AI Summary") {
		t.Fatalf("AI summary should be omitted on summarizer failure:\n%s", content)
	}
	if !strings.Contains(content, "summary_status: skipped (provider timeout)") {
		t.Fatalf("metadata should include skipped status:\n%s", content)
	}
}

func TestOpenAISummarizerAcceptsJSONInCodeFence(t *testing.T) {
	t.Parallel()

	clientHTTP := &http.Client{
		Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			return mustJSONResponse(http.StatusOK, map[string]any{
				"output": []map[string]any{
					{
						"type": "message",
						"content": []map[string]any{
							{
								"type": "output_text",
								"text": "```json\n{\"summary\":\"brief\",\"key_decisions\":[\"d1\"],\"action_items\":[\"a1\"],\"language\":\"en\"}\n```",
							},
						},
					},
				},
			}), nil
		}),
	}

	s := NewOpenAISummarizer(OpenAISummarizerConfig{
		APIKey:     "test-key",
		BaseURL:    "https://api.test",
		Model:      "gpt-test",
		HTTPClient: clientHTTP,
	})

	got, err := s.Summarize(context.Background(), sampleIssueData(), "en")
	if err != nil {
		t.Fatalf("Summarize error = %v, want nil", err)
	}
	if got.Summary != "brief" {
		t.Fatalf("Summary = %q, want brief", got.Summary)
	}
}

func TestBuildSourceTextIsCapped(t *testing.T) {
	t.Parallel()

	data := sampleIssueData()
	data.Description = strings.Repeat("a", maxSummarySourceChars+500)
	data.Thread = nil
	out := buildSourceText(data)
	if len(out) > maxSummarySourceChars {
		t.Fatalf("buildSourceText len = %d, want <= %d", len(out), maxSummarySourceChars)
	}
	if !strings.Contains(out, "[truncated]") {
		t.Fatalf("buildSourceText should include truncation marker: %s", fmt.Sprintf("%q", out[len(out)-80:]))
	}
}
