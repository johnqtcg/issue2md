package converter

import (
	"fmt"
	"strings"

	gh "github.com/johnqtcg/issue2md/internal/github"
)

func renderPRReviewsSection(data gh.IssueData, includeComments bool) string {
	var b strings.Builder

	b.WriteString("## Reviews\n")
	if !includeComments {
		b.WriteString("Reviews omitted (--include-comments=false).\n")
		return b.String()
	}
	if len(data.Reviews) == 0 {
		b.WriteString("- none\n")
		return b.String()
	}

	for _, review := range data.Reviews {
		fmt.Fprintf(&b, "- %s by %s at %s: %s\n", review.State, review.Author, review.CreatedAt, review.Body)
		for _, comment := range review.Comments {
			fmt.Fprintf(&b, "  - %s (%s): %s\n", comment.Author, comment.CreatedAt, comment.Body)
		}
	}

	return b.String()
}

func renderPRThreadSection(data gh.IssueData, includeComments bool) string {
	var b strings.Builder

	b.WriteString("## Discussion Thread\n")
	if !includeComments {
		b.WriteString("Comments omitted (--include-comments=false).\n")
		return b.String()
	}
	if len(data.Thread) == 0 {
		b.WriteString("- none\n")
		return b.String()
	}

	writeCommentList(&b, data.Thread, 0)
	return b.String()
}
