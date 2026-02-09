package converter

import (
	"fmt"
	"strings"

	gh "github.com/johnqtcg/issue2md/internal/github"
)

func renderDiscussionThreadSection(data gh.IssueData, includeComments bool) string {
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

	if data.Meta.IsAnswered {
		accepted, ok := resolveAcceptedAnswer(data.Thread, data.Meta.AcceptedAnswerID, data.Meta.AcceptedAnswerAuthor)
		if ok {
			b.WriteString("\n### Accepted Answer\n")
			fmt.Fprintf(&b, "- %s (%s): %s\n", accepted.Author, accepted.CreatedAt, accepted.Body)
		}
	}

	b.WriteString("\n### Replies\n")
	writeCommentList(&b, data.Thread, 0)
	return b.String()
}

func resolveAcceptedAnswer(nodes []gh.CommentNode, acceptedID, acceptedAuthor string) (gh.CommentNode, bool) {
	if acceptedID != "" {
		if accepted, ok := findAcceptedAnswerByID(nodes, acceptedID); ok {
			return accepted, true
		}
	}
	if acceptedAuthor != "" {
		return findAcceptedAnswerByAuthor(nodes, acceptedAuthor)
	}
	return gh.CommentNode{}, false
}

func findAcceptedAnswerByID(nodes []gh.CommentNode, id string) (gh.CommentNode, bool) {
	for _, node := range nodes {
		if node.ID == id {
			return node, true
		}
		if nested, ok := findAcceptedAnswerByID(node.Replies, id); ok {
			return nested, true
		}
	}
	return gh.CommentNode{}, false
}

func findAcceptedAnswerByAuthor(nodes []gh.CommentNode, author string) (gh.CommentNode, bool) {
	for _, node := range nodes {
		if node.Author == author {
			return node, true
		}
		if nested, ok := findAcceptedAnswerByAuthor(node.Replies, author); ok {
			return nested, true
		}
	}
	return gh.CommentNode{}, false
}

func writeCommentList(b *strings.Builder, comments []gh.CommentNode, depth int) {
	prefix := strings.Repeat("  ", depth)
	for _, comment := range comments {
		fmt.Fprintf(b, "%s- %s (%s): %s\n", prefix, comment.Author, comment.CreatedAt, comment.Body)
		if len(comment.Replies) > 0 {
			writeCommentList(b, comment.Replies, depth+1)
		}
	}
}
