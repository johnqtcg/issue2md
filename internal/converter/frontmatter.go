package converter

import (
	"fmt"
	"strings"

	gh "github.com/johnqtcg/issue2md/internal/github"
)

func renderFrontMatter(meta gh.Metadata) string {
	var b strings.Builder

	b.WriteString("---\n")
	fmt.Fprintf(&b, "type: %s\n", yamlQuote(string(meta.Type)))
	fmt.Fprintf(&b, "title: %s\n", yamlQuote(meta.Title))
	fmt.Fprintf(&b, "number: %d\n", meta.Number)
	fmt.Fprintf(&b, "state: %s\n", yamlQuote(meta.State))
	fmt.Fprintf(&b, "author: %s\n", yamlQuote(meta.Author))
	fmt.Fprintf(&b, "created_at: %s\n", yamlQuote(meta.CreatedAt))
	fmt.Fprintf(&b, "updated_at: %s\n", yamlQuote(meta.UpdatedAt))
	fmt.Fprintf(&b, "url: %s\n", yamlQuote(meta.URL))

	writeLabelList(&b, meta.Labels)

	switch meta.Type {
	case gh.ResourcePullRequest:
		fmt.Fprintf(&b, "merged: %t\n", meta.Merged)
		if meta.MergedAt != "" {
			fmt.Fprintf(&b, "merged_at: %s\n", yamlQuote(meta.MergedAt))
		}
		fmt.Fprintf(&b, "review_count: %d\n", meta.ReviewCount)
	case gh.ResourceDiscussion:
		if meta.Category != "" {
			fmt.Fprintf(&b, "category: %s\n", yamlQuote(meta.Category))
		}
		fmt.Fprintf(&b, "is_answered: %t\n", meta.IsAnswered)
		if meta.AcceptedAnswerAuthor != "" {
			fmt.Fprintf(&b, "accepted_answer_author: %s\n", yamlQuote(meta.AcceptedAnswerAuthor))
		}
	}

	b.WriteString("---\n\n")
	return b.String()
}

func writeLabelList(b *strings.Builder, labels []gh.Label) {
	if len(labels) == 0 {
		b.WriteString("labels: []\n")
		return
	}

	b.WriteString("labels:\n")
	for _, label := range labels {
		fmt.Fprintf(b, "  - %s\n", yamlQuote(label.Name))
	}
}

func yamlQuote(value string) string {
	escaped := strings.ReplaceAll(value, "'", "''")
	return "'" + escaped + "'"
}
