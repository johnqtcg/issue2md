package parser

import (
	"testing"

	gh "github.com/johnqtcg/issue2md/internal/github"
)

func TestParse(t *testing.T) {
	t.Parallel()

	tcs := []struct {
		name    string
		rawURL  string
		wantRef gh.ResourceRef
		wantErr bool
	}{
		{
			name:   "issue url",
			rawURL: "https://github.com/octo/repo/issues/123",
			wantRef: gh.ResourceRef{
				Owner:  "octo",
				Repo:   "repo",
				Number: 123,
				Type:   gh.ResourceIssue,
				URL:    "https://github.com/octo/repo/issues/123",
			},
		},
		{
			name:   "pr url with query and fragment",
			rawURL: "https://github.com/octo/repo/pull/42?foo=1#bar",
			wantRef: gh.ResourceRef{
				Owner:  "octo",
				Repo:   "repo",
				Number: 42,
				Type:   gh.ResourcePullRequest,
				URL:    "https://github.com/octo/repo/pull/42",
			},
		},
		{
			name:   "discussion url with trailing slash",
			rawURL: "https://github.com/octo/repo/discussions/99/",
			wantRef: gh.ResourceRef{
				Owner:  "octo",
				Repo:   "repo",
				Number: 99,
				Type:   gh.ResourceDiscussion,
				URL:    "https://github.com/octo/repo/discussions/99",
			},
		},
		{
			name:    "invalid host",
			rawURL:  "https://gitlab.com/octo/repo/issues/1",
			wantErr: true,
		},
		{
			name:    "unsupported path",
			rawURL:  "https://github.com/octo/repo/commits/1",
			wantErr: true,
		},
		{
			name:    "issue subresource path is rejected",
			rawURL:  "https://github.com/octo/repo/issues/1/comments",
			wantErr: true,
		},
		{
			name:    "pr subresource path is rejected",
			rawURL:  "https://github.com/octo/repo/pull/2/files",
			wantErr: true,
		},
		{
			name:    "invalid number",
			rawURL:  "https://github.com/octo/repo/issues/not-number",
			wantErr: true,
		},
		{
			name:    "missing repo",
			rawURL:  "https://github.com/octo//issues/1",
			wantErr: true,
		},
	}

	p := New()
	for _, tc := range tcs {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := p.Parse(tc.rawURL)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("Parse(%q) error = nil, want error", tc.rawURL)
				}
				return
			}
			if err != nil {
				t.Fatalf("Parse(%q) error = %v, want nil", tc.rawURL, err)
			}
			if got != tc.wantRef {
				t.Fatalf("Parse(%q) = %#v, want %#v", tc.rawURL, got, tc.wantRef)
			}
		})
	}
}
