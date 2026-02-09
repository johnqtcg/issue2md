package converter

import (
	"strings"
	"testing"
)

func TestRenderFrontMatterRequiredFields(t *testing.T) {
	t.Parallel()

	out := renderFrontMatter(sampleIssueData().Meta)
	required := []string{
		"type: 'issue'",
		"title: 'Issue: Panic on nil config'",
		"number: 123",
		"state: 'open'",
		"author: 'alice'",
		"created_at: '2026-01-01T10:00:00Z'",
		"updated_at: '2026-01-02T11:00:00Z'",
		"url: 'https://github.com/octo/repo/issues/123'",
		"labels:",
		"  - 'bug'",
	}
	for _, piece := range required {
		if !strings.Contains(out, piece) {
			t.Fatalf("front matter missing %q\n%s", piece, out)
		}
	}
}

func TestRenderFrontMatterOptionalFieldsByType(t *testing.T) {
	t.Parallel()

	tcs := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:  "pr optional fields",
			input: renderFrontMatter(samplePRData().Meta),
			expected: []string{
				"merged: true",
				"merged_at: '2026-01-04T09:30:00Z'",
				"review_count: 2",
			},
		},
		{
			name:  "discussion optional fields",
			input: renderFrontMatter(sampleDiscussionData().Meta),
			expected: []string{
				"category: 'Q&A'",
				"is_answered: true",
				"accepted_answer_author: 'mentor'",
			},
		},
	}

	for _, tc := range tcs {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			for _, piece := range tc.expected {
				if !strings.Contains(tc.input, piece) {
					t.Fatalf("front matter missing %q\n%s", piece, tc.input)
				}
			}
		})
	}
}

func TestRenderFrontMatterPreservesDatetimeString(t *testing.T) {
	t.Parallel()

	meta := sampleIssueData().Meta
	meta.CreatedAt = "2026-01-01T10:00:00+08:00"
	meta.UpdatedAt = "2026-01-02T11:00:00-07:00"

	out := renderFrontMatter(meta)
	if !strings.Contains(out, "created_at: '2026-01-01T10:00:00+08:00'") {
		t.Fatalf("created_at is not preserved:\n%s", out)
	}
	if !strings.Contains(out, "updated_at: '2026-01-02T11:00:00-07:00'") {
		t.Fatalf("updated_at is not preserved:\n%s", out)
	}
}
