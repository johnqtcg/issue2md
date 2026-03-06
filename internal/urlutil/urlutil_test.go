package urlutil

import (
	"strings"
	"testing"
)

func TestValidatePublicHTTPSURL(t *testing.T) {
	t.Parallel()

	tcs := []struct {
		name    string
		rawURL  string
		wantErr string
	}{
		{name: "valid public https", rawURL: "https://api.example.com/v1"},
		{name: "reject non https", rawURL: "http://api.example.com", wantErr: "https"},
		{name: "reject empty host", rawURL: "https:///path", wantErr: "host is empty"},
		{name: "reject userinfo", rawURL: "https://user:pass@api.example.com", wantErr: "userinfo"},
		{name: "reject private ip", rawURL: "https://127.0.0.1/api", wantErr: "private ip"},
	}

	for _, tc := range tcs {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := ValidatePublicHTTPSURL(tc.rawURL, "test endpoint")
			if tc.wantErr == "" {
				if err != nil {
					t.Fatalf("ValidatePublicHTTPSURL(%q) error = %v, want nil", tc.rawURL, err)
				}
				return
			}
			if err == nil {
				t.Fatalf("ValidatePublicHTTPSURL(%q) error = nil, want contains %q", tc.rawURL, tc.wantErr)
			}
			if got := err.Error(); got == "" || !strings.Contains(got, tc.wantErr) {
				t.Fatalf("ValidatePublicHTTPSURL(%q) error = %q, want contains %q", tc.rawURL, got, tc.wantErr)
			}
		})
	}
}

func TestResolvePublicHTTPSURLUsesDefault(t *testing.T) {
	t.Parallel()

	got, err := ResolvePublicHTTPSURL("  ", "https://api.example.com/graphql", "graphql endpoint")
	if err != nil {
		t.Fatalf("ResolvePublicHTTPSURL default error = %v, want nil", err)
	}
	if got != "https://api.example.com/graphql" {
		t.Fatalf("ResolvePublicHTTPSURL default = %q, want %q", got, "https://api.example.com/graphql")
	}
}
