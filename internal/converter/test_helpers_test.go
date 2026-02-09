package converter

import (
	"context"

	gh "github.com/johnqtcg/issue2md/internal/github"
)

func sampleIssueData() gh.IssueData {
	return gh.IssueData{
		Meta: gh.Metadata{
			Type:      gh.ResourceIssue,
			Title:     "Issue: Panic on nil config",
			Number:    123,
			State:     "open",
			Author:    "alice",
			CreatedAt: "2026-01-01T10:00:00Z",
			UpdatedAt: "2026-01-02T11:00:00Z",
			URL:       "https://github.com/octo/repo/issues/123",
			Labels: []gh.Label{
				{Name: "bug"},
				{Name: "help wanted"},
			},
		},
		Description: "App panics when config is nil.\n\n![image](https://example.com/a.png)",
		Timeline: []gh.TimelineEvent{
			{EventType: "opened", Actor: "alice", CreatedAt: "2026-01-01T10:00:00Z", Details: "Issue opened"},
			{EventType: "labeled", Actor: "bot", CreatedAt: "2026-01-01T10:30:00Z", Details: "bug"},
			{EventType: "assigned", Actor: "maintainer", CreatedAt: "2026-01-01T11:00:00Z", Details: "assigned to maintainer"},
		},
		Thread: []gh.CommentNode{
			{
				ID:        "c1",
				Author:    "bob",
				Body:      "I can reproduce this.",
				CreatedAt: "2026-01-01T12:00:00Z",
				Replies: []gh.CommentNode{
					{
						ID:        "c1-r1",
						Author:    "alice",
						Body:      "Thanks, investigating.",
						CreatedAt: "2026-01-01T12:30:00Z",
					},
				},
			},
			{
				ID:        "c2",
				Author:    "carol",
				Body:      "Fixed in #124?",
				CreatedAt: "2026-01-01T13:00:00Z",
			},
		},
	}
}

func samplePRData() gh.IssueData {
	return gh.IssueData{
		Meta: gh.Metadata{
			Type:        gh.ResourcePullRequest,
			Title:       "PR: Fix nil config panic",
			Number:      124,
			State:       "closed",
			Author:      "alice",
			CreatedAt:   "2026-01-03T09:00:00Z",
			UpdatedAt:   "2026-01-04T10:00:00Z",
			URL:         "https://github.com/octo/repo/pull/124",
			Labels:      []gh.Label{{Name: "bugfix"}},
			Merged:      true,
			MergedAt:    "2026-01-04T09:30:00Z",
			ReviewCount: 2,
		},
		Description: "This PR adds a nil check.",
		Reviews: []gh.ReviewData{
			{
				ID:        "r1",
				State:     "APPROVED",
				Author:    "bob",
				Body:      "Looks good.",
				CreatedAt: "2026-01-03T12:00:00Z",
				Comments: []gh.CommentNode{
					{
						ID:        "r1-c1",
						Author:    "bob",
						Body:      "Please add test.",
						CreatedAt: "2026-01-03T12:10:00Z",
					},
				},
			},
			{
				ID:        "r2",
				State:     "CHANGES_REQUESTED",
				Author:    "carol",
				Body:      "Need edge case coverage.",
				CreatedAt: "2026-01-03T13:00:00Z",
			},
		},
		Thread: []gh.CommentNode{
			{
				ID:        "pr-thread-1",
				Author:    "dave",
				Body:      "Great improvement.",
				CreatedAt: "2026-01-03T14:00:00Z",
			},
		},
	}
}

func sampleDiscussionData() gh.IssueData {
	return gh.IssueData{
		Meta: gh.Metadata{
			Type:                 gh.ResourceDiscussion,
			Title:                "How to configure issue2md?",
			Number:               88,
			State:                "open",
			Author:               "dora",
			CreatedAt:            "2026-01-05T09:00:00Z",
			UpdatedAt:            "2026-01-05T10:00:00Z",
			URL:                  "https://github.com/octo/repo/discussions/88",
			Category:             "Q&A",
			IsAnswered:           true,
			AcceptedAnswerID:     "d2",
			AcceptedAnswerAuthor: "mentor",
		},
		Description: "What's the best config for tokens?",
		Thread: []gh.CommentNode{
			{
				ID:        "d1",
				Author:    "dora",
				Body:      "Any best practice for setup?",
				CreatedAt: "2026-01-05T09:10:00Z",
			},
			{
				ID:        "d2",
				Author:    "mentor",
				Body:      "Set GITHUB_TOKEN and OPENAI_API_KEY in your shell.",
				CreatedAt: "2026-01-05T09:15:00Z",
				Replies: []gh.CommentNode{
					{
						ID:        "d2-r1",
						Author:    "dora",
						Body:      "Thanks, this worked.",
						CreatedAt: "2026-01-05T09:20:00Z",
					},
				},
			},
		},
	}
}

func fixedSummary() Summary {
	return Summary{
		Summary:      "The thread discusses root cause and fix.",
		KeyDecisions: []string{"Use nil guard before dereference.", "Backfill regression tests."},
		ActionItems:  []string{"Release v1.0.1.", "Update documentation."},
		Language:     "en",
		Status:       "ok",
	}
}

type stubSummarizer struct {
	summary  Summary
	err      error
	lastLang string
}

func (s *stubSummarizer) Summarize(_ context.Context, _ gh.IssueData, lang string) (Summary, error) {
	s.lastLang = lang
	if s.err != nil {
		return Summary{}, s.err
	}
	return s.summary, nil
}
