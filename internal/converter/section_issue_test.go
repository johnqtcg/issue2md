package converter

import (
	"strings"
	"testing"
)

func TestRenderIssueTimelineSection(t *testing.T) {
	t.Parallel()

	out := renderIssueTimelineSection(sampleIssueData())
	expected := []string{
		"## Timeline",
		"- 2026-01-01T10:00:00Z | opened | alice | Issue opened",
		"- 2026-01-01T10:30:00Z | labeled | bot | bug",
	}
	for _, piece := range expected {
		if !strings.Contains(out, piece) {
			t.Fatalf("timeline missing %q\n%s", piece, out)
		}
	}
}

func TestRenderIssueThreadSectionIncludeCommentsOption(t *testing.T) {
	t.Parallel()

	withComments := renderIssueThreadSection(sampleIssueData(), true)
	if !strings.Contains(withComments, "I can reproduce this.") {
		t.Fatalf("thread missing comment when includeComments=true:\n%s", withComments)
	}
	if !strings.Contains(withComments, "Thanks, investigating.") {
		t.Fatalf("thread missing nested reply when includeComments=true:\n%s", withComments)
	}

	withoutComments := renderIssueThreadSection(sampleIssueData(), false)
	if strings.Contains(withoutComments, "I can reproduce this.") {
		t.Fatalf("thread should omit comments when includeComments=false:\n%s", withoutComments)
	}
	if !strings.Contains(withoutComments, "Comments omitted (--include-comments=false).") {
		t.Fatalf("thread should include omitted note:\n%s", withoutComments)
	}
}
