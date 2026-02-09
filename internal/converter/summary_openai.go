package converter

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	gh "github.com/johnqtcg/issue2md/internal/github"
)

const (
	defaultOpenAIBaseURL  = "https://api.openai.com"
	defaultOpenAIModel    = "gpt-5-mini"
	maxSummarySourceChars = 12000
	summaryTruncatedNote  = "\n[truncated]\n"
)

// Summary holds the normalized AI summary payload for markdown rendering.
type Summary struct {
	Summary      string
	KeyDecisions []string
	ActionItems  []string
	Language     string
	Status       string
	Reason       string
}

// Summarizer defines the AI summary capability used by renderer.
type Summarizer interface {
	Summarize(ctx context.Context, data gh.IssueData, lang string) (Summary, error)
}

// OpenAISummarizerConfig configures OpenAI Responses API integration.
type OpenAISummarizerConfig struct {
	APIKey     string
	BaseURL    string
	Model      string
	HTTPClient *http.Client
}

type openAISummarizer struct {
	httpClient *http.Client
	endpoint   string
	model      string
	apiKey     string
}

type openAIResponseEnvelope struct {
	Output []struct {
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
	} `json:"output"`
}

type openAISummaryPayload struct {
	Summary      string   `json:"summary"`
	KeyDecisions []string `json:"key_decisions"`
	ActionItems  []string `json:"action_items"`
	Language     string   `json:"language"`
}

// NewOpenAISummarizer creates a Summarizer backed by OpenAI Responses API.
func NewOpenAISummarizer(cfg OpenAISummarizerConfig) Summarizer {
	httpClient := cfg.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 45 * time.Second}
	}

	model := cfg.Model
	if model == "" {
		model = defaultOpenAIModel
	}

	return &openAISummarizer{
		httpClient: httpClient,
		endpoint:   buildResponsesEndpoint(cfg.BaseURL),
		model:      model,
		apiKey:     cfg.APIKey,
	}
}

func (s *openAISummarizer) Summarize(ctx context.Context, data gh.IssueData, lang string) (Summary, error) {
	if s.apiKey == "" {
		return Summary{}, fmt.Errorf("openai api key is empty")
	}

	targetLang := resolveSummaryLanguage(lang, data)
	payload := map[string]any{
		"model": s.model,
		"input": buildSummaryPrompt(data, targetLang),
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return Summary{}, fmt.Errorf("marshal summary request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.endpoint, bytes.NewReader(body))
	if err != nil {
		return Summary{}, fmt.Errorf("create summary request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+s.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return Summary{}, fmt.Errorf("execute summary request: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode >= http.StatusBadRequest {
		msg, readErr := io.ReadAll(io.LimitReader(resp.Body, 16*1024))
		if readErr != nil {
			return Summary{}, fmt.Errorf("read summary error response: %w", readErr)
		}
		return Summary{}, fmt.Errorf("summary request failed with status %d: %s", resp.StatusCode, strings.TrimSpace(string(msg)))
	}

	var envelope openAIResponseEnvelope
	if err := json.NewDecoder(resp.Body).Decode(&envelope); err != nil {
		return Summary{}, fmt.Errorf("decode summary response: %w", err)
	}

	text, err := extractSummaryText(envelope)
	if err != nil {
		return Summary{}, fmt.Errorf("extract summary text: %w", err)
	}
	jsonText, err := normalizeSummaryJSON(text)
	if err != nil {
		return Summary{}, fmt.Errorf("normalize summary json: %w", err)
	}

	var out openAISummaryPayload
	if err := json.Unmarshal([]byte(jsonText), &out); err != nil {
		return Summary{}, fmt.Errorf("decode summary payload: %w", err)
	}
	if out.Summary == "" {
		return Summary{}, fmt.Errorf("summary payload missing summary field")
	}
	if out.Language == "" {
		out.Language = targetLang
	}

	return Summary{
		Summary:      out.Summary,
		KeyDecisions: out.KeyDecisions,
		ActionItems:  out.ActionItems,
		Language:     out.Language,
		Status:       "ok",
	}, nil
}

func buildResponsesEndpoint(baseURL string) string {
	if baseURL == "" {
		return defaultOpenAIBaseURL + "/v1/responses"
	}
	trimmed := strings.TrimRight(baseURL, "/")
	if strings.HasSuffix(trimmed, "/v1") {
		return trimmed + "/responses"
	}
	return trimmed + "/v1/responses"
}

func buildSummaryPrompt(data gh.IssueData, lang string) string {
	return fmt.Sprintf(
		"Summarize the following GitHub discussion archive.\nLanguage: %s\nReturn strict JSON only (no markdown) with keys: summary (string), key_decisions (string array), action_items (string array), language (string).\n\nSource:\n%s",
		lang,
		buildSourceText(data),
	)
}

func buildSourceText(data gh.IssueData) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Title: %s\n", data.Meta.Title)
	fmt.Fprintf(&b, "Type: %s\n", data.Meta.Type)
	fmt.Fprintf(&b, "Description:\n%s\n", data.Description)
	appendThreadText(&b, data.Thread)
	return capSummarySourceLength(b.String())
}

func appendThreadText(b *strings.Builder, nodes []gh.CommentNode) {
	for _, node := range nodes {
		fmt.Fprintf(b, "\nComment by %s:\n%s\n", node.Author, node.Body)
		appendThreadText(b, node.Replies)
	}
}

func extractSummaryText(envelope openAIResponseEnvelope) (string, error) {
	for _, output := range envelope.Output {
		for _, content := range output.Content {
			if content.Type == "output_text" && content.Text != "" {
				return content.Text, nil
			}
		}
	}
	return "", fmt.Errorf("no output_text found in response")
}

func normalizeSummaryJSON(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", fmt.Errorf("summary output is empty")
	}

	if strings.HasPrefix(trimmed, "```") {
		lines := strings.Split(trimmed, "\n")
		if len(lines) > 1 {
			lines = lines[1:]
			if len(lines) > 0 && strings.HasPrefix(strings.TrimSpace(lines[len(lines)-1]), "```") {
				lines = lines[:len(lines)-1]
			}
			trimmed = strings.TrimSpace(strings.Join(lines, "\n"))
		}
	}

	start := strings.Index(trimmed, "{")
	end := strings.LastIndex(trimmed, "}")
	if start == -1 || end == -1 || end < start {
		return "", fmt.Errorf("summary output does not contain a json object")
	}

	jsonText := strings.TrimSpace(trimmed[start : end+1])
	if !json.Valid([]byte(jsonText)) {
		return "", fmt.Errorf("summary output is not valid json")
	}
	return jsonText, nil
}

func capSummarySourceLength(source string) string {
	runes := []rune(source)
	if len(runes) <= maxSummarySourceChars {
		return source
	}

	noteRunes := []rune(summaryTruncatedNote)
	cut := maxSummarySourceChars - len(noteRunes)
	if cut < 0 {
		cut = 0
	}
	return string(runes[:cut]) + summaryTruncatedNote
}

func resolveSummaryLanguage(override string, data gh.IssueData) string {
	if override != "" {
		return override
	}

	source := data.Meta.Title + "\n" + data.Description
	for _, r := range source {
		if isLikelyChineseRune(r) {
			return "zh"
		}
	}
	return "en"
}

func isLikelyChineseRune(r rune) bool {
	return r >= 0x4E00 && r <= 0x9FFF
}
