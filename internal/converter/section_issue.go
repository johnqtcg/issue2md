package converter

import (
	"fmt"
	"strings"

	gh "github.com/johnqtcg/issue2md/internal/github"
)

func renderIssueTimelineSection(data gh.IssueData) string {
	var b strings.Builder

	b.WriteString("## Timeline\n")
	if len(data.Timeline) == 0 {
		b.WriteString("- none\n")
		return b.String()
	}

	for _, event := range data.Timeline {
		fmt.Fprintf(&b, "- %s | %s | %s | %s\n", event.CreatedAt, event.EventType, event.Actor, event.Details)
	}

	return b.String()
}

func renderIssueThreadSection(data gh.IssueData, includeComments bool) string {
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
