package github

import (
	"errors"
	"net/http"
	"testing"
)

func TestIsAuthError(t *testing.T) {
	t.Parallel()

	tcs := []struct {
		err  error
		name string
		want bool
	}{
		{name: "401 is auth", err: NewStatusError(http.StatusUnauthorized, errors.New("bad credentials"), nil), want: true},
		{name: "text 401 bad credentials is auth", err: errors.New("http status 401: bad credentials"), want: true},
		{name: "403 forbidden is auth", err: NewStatusError(http.StatusForbidden, errors.New("forbidden"), nil), want: true},
		{name: "403 rate limit header is not auth", err: NewStatusError(http.StatusForbidden, errors.New("forbidden"), http.Header{"X-RateLimit-Reset": []string{"1893456000"}}), want: false},
		{name: "429 is not auth", err: NewStatusError(http.StatusTooManyRequests, errors.New("rate limit"), nil), want: false},
		{name: "text unauthorized is auth", err: errors.New("unauthorized"), want: true},
		{name: "text rate limit is not auth", err: errors.New("status 403: API rate limit exceeded"), want: false},
	}

	for _, tc := range tcs {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := IsAuthError(tc.err); got != tc.want {
				t.Fatalf("IsAuthError(%v) = %v, want %v", tc.err, got, tc.want)
			}
		})
	}
}

func TestIsRateLimitError(t *testing.T) {
	t.Parallel()

	tcs := []struct {
		err  error
		name string
		want bool
	}{
		{name: "429 status", err: NewStatusError(http.StatusTooManyRequests, errors.New("too many requests"), nil), want: true},
		{name: "403 reset header", err: NewStatusError(http.StatusForbidden, errors.New("forbidden"), http.Header{"X-RateLimit-Reset": []string{"1893456000"}}), want: true},
		{name: "403 retry after", err: NewStatusError(http.StatusForbidden, errors.New("forbidden"), http.Header{"Retry-After": []string{"7"}}), want: true},
		{name: "403 forbidden", err: NewStatusError(http.StatusForbidden, errors.New("forbidden"), nil), want: false},
		{name: "text fallback", err: errors.New("API rate limit exceeded"), want: true},
	}

	for _, tc := range tcs {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := IsRateLimitError(tc.err); got != tc.want {
				t.Fatalf("IsRateLimitError(%v) = %v, want %v", tc.err, got, tc.want)
			}
		})
	}
}
