package converter

import (
	"context"
	"os"
	"strings"
	"testing"
)

func TestRendererSectionOrder(t *testing.T) {
	t.Parallel()

	r := NewRenderer(&stubSummarizer{summary: fixedSummary()})
	out, err := r.Render(context.Background(), sampleIssueData(), RenderOptions{
		IncludeComments: true,
		IncludeSummary:  true,
	})
	if err != nil {
		t.Fatalf("Render error = %v, want nil", err)
	}
	content := string(out)

	ordered := []string{
		"# Issue: Panic on nil config",
		"## Metadata",
		"## AI Summary",
		"## Original Description",
		"## Timeline",
		"## Discussion Thread",
		"## References",
	}

	last := -1
	for _, piece := range ordered {
		idx := strings.Index(content, piece)
		if idx < 0 {
			t.Fatalf("missing section %q\n%s", piece, content)
		}
		if idx <= last {
			t.Fatalf("section order incorrect around %q\n%s", piece, content)
		}
		last = idx
	}
}

func TestRendererIncludeCommentsOption(t *testing.T) {
	t.Parallel()

	r := NewRenderer(&stubSummarizer{summary: fixedSummary()})
	out, err := r.Render(context.Background(), samplePRData(), RenderOptions{
		IncludeComments: false,
		IncludeSummary:  true,
	})
	if err != nil {
		t.Fatalf("Render error = %v, want nil", err)
	}

	content := string(out)
	if strings.Contains(content, "Looks good.") {
		t.Fatalf("PR reviews should be omitted when include-comments=false:\n%s", content)
	}
	if !strings.Contains(content, "Reviews omitted (--include-comments=false).") {
		t.Fatalf("PR output should include omitted note:\n%s", content)
	}
}

func TestRendererGoldenIssuePRDiscussion(t *testing.T) {
	t.Parallel()

	tcs := []struct {
		name   string
		data   func() string
		golden string
	}{
		{
			name: "issue",
			data: func() string {
				out, err := NewRenderer(&stubSummarizer{summary: fixedSummary()}).Render(context.Background(), sampleIssueData(), RenderOptions{IncludeComments: true, IncludeSummary: true})
				if err != nil {
					t.Fatalf("Render issue error = %v", err)
				}
				return string(out)
			},
			golden: "testdata/issue.golden.md",
		},
		{
			name: "pr",
			data: func() string {
				out, err := NewRenderer(&stubSummarizer{summary: fixedSummary()}).Render(context.Background(), samplePRData(), RenderOptions{IncludeComments: true, IncludeSummary: true})
				if err != nil {
					t.Fatalf("Render pr error = %v", err)
				}
				return string(out)
			},
			golden: "testdata/pr.golden.md",
		},
		{
			name: "discussion",
			data: func() string {
				out, err := NewRenderer(&stubSummarizer{summary: fixedSummary()}).Render(context.Background(), sampleDiscussionData(), RenderOptions{IncludeComments: true, IncludeSummary: true})
				if err != nil {
					t.Fatalf("Render discussion error = %v", err)
				}
				return string(out)
			},
			golden: "testdata/discussion.golden.md",
		},
	}

	for _, tc := range tcs {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := tc.data()
			wantBytes, err := os.ReadFile(tc.golden)
			if err != nil {
				t.Fatalf("ReadFile(%s) error = %v", tc.golden, err)
			}
			want := string(wantBytes)
			if got != want {
				t.Fatalf("golden mismatch for %s\n--- got ---\n%s\n--- want ---\n%s", tc.name, got, want)
			}
		})
	}
}
