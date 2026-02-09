package converter

import (
	"strings"
	"testing"
)

func TestRenderPRReviewsSection(t *testing.T) {
	t.Parallel()

	out := renderPRReviewsSection(samplePRData(), true)
	expected := []string{
		"## Reviews",
		"- APPROVED by bob at 2026-01-03T12:00:00Z: Looks good.",
		"- CHANGES_REQUESTED by carol at 2026-01-03T13:00:00Z: Need edge case coverage.",
		"Please add test.",
	}
	for _, piece := range expected {
		if !strings.Contains(out, piece) {
			t.Fatalf("reviews missing %q\n%s", piece, out)
		}
	}
	if strings.Contains(out, "diff --git") {
		t.Fatalf("reviews should not include diff details:\n%s", out)
	}
}

func TestRenderPRSectionsIncludeCommentsOption(t *testing.T) {
	t.Parallel()

	withComments := renderPRReviewsSection(samplePRData(), true)
	if !strings.Contains(withComments, "Please add test.") {
		t.Fatalf("review thread comment missing when includeComments=true:\n%s", withComments)
	}

	withoutComments := renderPRReviewsSection(samplePRData(), false)
	if !strings.Contains(withoutComments, "Reviews omitted (--include-comments=false).") {
		t.Fatalf("reviews should include omitted note:\n%s", withoutComments)
	}
	if strings.Contains(withoutComments, "Looks good.") {
		t.Fatalf("reviews should be omitted when includeComments=false:\n%s", withoutComments)
	}
}
