package converter

import (
	"strings"
	"testing"

	gh "github.com/johnqtcg/issue2md/internal/github"
)

func TestRenderDiscussionThreadSection(t *testing.T) {
	t.Parallel()

	out := renderDiscussionThreadSection(sampleDiscussionData(), true)
	expected := []string{
		"## Discussion Thread",
		"### Accepted Answer",
		"Set GITHUB_TOKEN and OPENAI_API_KEY in your shell.",
		"### Replies",
		"Thanks, this worked.",
	}
	for _, piece := range expected {
		if !strings.Contains(out, piece) {
			t.Fatalf("discussion thread missing %q\n%s", piece, out)
		}
	}
}

func TestRenderDiscussionThreadIncludeCommentsOption(t *testing.T) {
	t.Parallel()

	out := renderDiscussionThreadSection(sampleDiscussionData(), false)
	if !strings.Contains(out, "Comments omitted (--include-comments=false).") {
		t.Fatalf("discussion thread should include omitted note:\n%s", out)
	}
	if strings.Contains(out, "Any best practice for setup?") {
		t.Fatalf("discussion comments should be omitted:\n%s", out)
	}
}

func TestRenderDiscussionThreadSectionUsesAcceptedAnswerID(t *testing.T) {
	t.Parallel()

	data := sampleDiscussionData()
	data.Meta.AcceptedAnswerAuthor = "mentor"
	data.Meta.AcceptedAnswerID = "d3"
	data.Thread = []gh.CommentNode{
		{ID: "d2", Author: "mentor", Body: "not accepted", CreatedAt: "2026-01-05T09:15:00Z"},
		{ID: "d3", Author: "mentor", Body: "accepted answer", CreatedAt: "2026-01-05T09:16:00Z"},
	}

	out := renderDiscussionThreadSection(data, true)
	if !strings.Contains(out, "accepted answer") {
		t.Fatalf("discussion should include accepted answer by id:\n%s", out)
	}
	if strings.Contains(out, "### Accepted Answer\n- mentor (2026-01-05T09:15:00Z): not accepted") {
		t.Fatalf("discussion used author-based matching instead of accepted answer id:\n%s", out)
	}
}
