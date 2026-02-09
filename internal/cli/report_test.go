package cli

import (
	"strings"
	"testing"

	gh "github.com/johnqtcg/issue2md/internal/github"
)

func TestBuildSummaryCounts(t *testing.T) {
	t.Parallel()

	items := []ItemResult{
		{URL: "u1", ResourceType: gh.ResourceIssue, Status: StatusOK},
		{URL: "u2", ResourceType: gh.ResourcePullRequest, Status: StatusFailed, Reason: "timeout"},
		{URL: "u3", ResourceType: gh.ResourceDiscussion, Status: StatusOK},
	}

	got := BuildSummary(items)
	if got.Total != 3 {
		t.Fatalf("Total = %d, want 3", got.Total)
	}
	if got.Succeeded != 2 {
		t.Fatalf("Succeeded = %d, want 2", got.Succeeded)
	}
	if got.Failed != 1 {
		t.Fatalf("Failed = %d, want 1", got.Failed)
	}
}

func TestFormatSummaryContainsStatusAndFailureEntries(t *testing.T) {
	t.Parallel()

	s := RunSummary{
		Total:     2,
		Succeeded: 1,
		Failed:    1,
		Items: []ItemResult{
			{URL: "https://github.com/octo/repo/issues/1", ResourceType: gh.ResourceIssue, Status: StatusOK},
			{
				URL:          "https://github.com/octo/repo/pull/2",
				ResourceType: gh.ResourcePullRequest,
				Status:       StatusFailed,
				Reason:       "network timeout",
			},
		},
	}

	out := FormatSummary(s)
	expected := []string{
		"OK total=2 succeeded=1 failed=1",
		"FAILED url=https://github.com/octo/repo/pull/2 type=pull_request reason=network timeout",
	}
	for _, piece := range expected {
		if !strings.Contains(out, piece) {
			t.Fatalf("summary missing %q\n%s", piece, out)
		}
	}
}
