package converter

import (
	"context"
	"fmt"
	"strings"

	gh "github.com/johnqtcg/issue2md/internal/github"
)

// RenderOptions controls markdown rendering behavior.
type RenderOptions struct {
	IncludeComments bool
	IncludeSummary  bool
	Lang            string
}

// Renderer converts normalized GitHub data into markdown output.
type Renderer interface {
	Render(ctx context.Context, data gh.IssueData, opts RenderOptions) ([]byte, error)
}

type renderer struct {
	summarizer Summarizer
}

// NewRenderer creates a markdown renderer instance.
func NewRenderer(summarizer Summarizer) Renderer {
	return &renderer{summarizer: summarizer}
}

func (r *renderer) Render(ctx context.Context, data gh.IssueData, opts RenderOptions) ([]byte, error) {
	if data.Meta.Type == "" {
		return nil, fmt.Errorf("render markdown: missing resource type")
	}

	var (
		summary       Summary
		summaryStatus string
	)

	if opts.IncludeSummary && r.summarizer != nil {
		targetLang := resolveSummaryLanguage(opts.Lang, data)
		got, err := r.summarizer.Summarize(ctx, data, targetLang)
		switch {
		case err != nil:
			summaryStatus = fmt.Sprintf("skipped (%s)", err.Error())
		case got.Status == "skipped":
			reason := got.Reason
			if reason == "" {
				reason = "summary unavailable"
			}
			summaryStatus = fmt.Sprintf("skipped (%s)", reason)
		default:
			summary = got
		}
	}

	var b strings.Builder
	b.WriteString(renderFrontMatter(data.Meta))
	fmt.Fprintf(&b, "# %s\n\n", data.Meta.Title)
	b.WriteString(renderMetadataSection(data.Meta, summaryStatus))

	if summary.Summary != "" {
		b.WriteString("\n")
		b.WriteString(renderSummarySection(summary))
	}

	b.WriteString("\n## Original Description\n\n")
	if strings.TrimSpace(data.Description) == "" {
		b.WriteString("(empty)\n")
	} else {
		b.WriteString(data.Description)
		b.WriteString("\n")
	}

	switch data.Meta.Type {
	case gh.ResourceIssue:
		b.WriteString("\n")
		b.WriteString(renderIssueTimelineSection(data))
		b.WriteString("\n")
		b.WriteString(renderIssueThreadSection(data, opts.IncludeComments))
	case gh.ResourcePullRequest:
		b.WriteString("\n")
		b.WriteString(renderPRReviewsSection(data, opts.IncludeComments))
		b.WriteString("\n")
		b.WriteString(renderPRThreadSection(data, opts.IncludeComments))
	case gh.ResourceDiscussion:
		b.WriteString("\n")
		b.WriteString(renderDiscussionThreadSection(data, opts.IncludeComments))
	default:
		return nil, fmt.Errorf("render markdown: unsupported resource type %q", data.Meta.Type)
	}

	b.WriteString("\n## References\n")
	fmt.Fprintf(&b, "- Original URL: %s\n", data.Meta.URL)

	return []byte(b.String()), nil
}

func renderMetadataSection(meta gh.Metadata, summaryStatus string) string {
	var b strings.Builder

	b.WriteString("## Metadata\n")
	fmt.Fprintf(&b, "- type: %s\n", meta.Type)
	fmt.Fprintf(&b, "- number: %d\n", meta.Number)
	fmt.Fprintf(&b, "- state: %s\n", meta.State)
	fmt.Fprintf(&b, "- author: %s\n", meta.Author)
	fmt.Fprintf(&b, "- created_at: %s\n", meta.CreatedAt)
	fmt.Fprintf(&b, "- updated_at: %s\n", meta.UpdatedAt)
	fmt.Fprintf(&b, "- url: %s\n", meta.URL)
	fmt.Fprintf(&b, "- labels: %s\n", joinLabels(meta.Labels))

	if meta.Type == gh.ResourcePullRequest {
		fmt.Fprintf(&b, "- merged: %t\n", meta.Merged)
		if meta.MergedAt != "" {
			fmt.Fprintf(&b, "- merged_at: %s\n", meta.MergedAt)
		}
		fmt.Fprintf(&b, "- review_count: %d\n", meta.ReviewCount)
	}
	if meta.Type == gh.ResourceDiscussion {
		if meta.Category != "" {
			fmt.Fprintf(&b, "- category: %s\n", meta.Category)
		}
		fmt.Fprintf(&b, "- is_answered: %t\n", meta.IsAnswered)
		if meta.AcceptedAnswerAuthor != "" {
			fmt.Fprintf(&b, "- accepted_answer_author: %s\n", meta.AcceptedAnswerAuthor)
		}
	}
	if summaryStatus != "" {
		fmt.Fprintf(&b, "- summary_status: %s\n", summaryStatus)
	}

	return b.String()
}

func renderSummarySection(summary Summary) string {
	var b strings.Builder

	b.WriteString("## AI Summary\n\n")
	b.WriteString("### Summary\n")
	b.WriteString(summary.Summary)
	b.WriteString("\n\n")

	b.WriteString("### Key Decisions\n")
	if len(summary.KeyDecisions) == 0 {
		b.WriteString("- none\n")
	} else {
		for _, item := range summary.KeyDecisions {
			fmt.Fprintf(&b, "- %s\n", item)
		}
	}
	b.WriteString("\n")

	b.WriteString("### Action Items\n")
	if len(summary.ActionItems) == 0 {
		b.WriteString("- none\n")
	} else {
		for _, item := range summary.ActionItems {
			fmt.Fprintf(&b, "- %s\n", item)
		}
	}

	return b.String()
}

func joinLabels(labels []gh.Label) string {
	if len(labels) == 0 {
		return "none"
	}
	parts := make([]string, 0, len(labels))
	for _, label := range labels {
		parts = append(parts, label.Name)
	}
	return strings.Join(parts, ", ")
}
