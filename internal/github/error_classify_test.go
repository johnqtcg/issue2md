package github

import (
	"errors"
	"net/http"
	"testing"
)

func TestIsAuthError(t *testing.T) {
	t.Parallel()

	tcs := []struct {
		name string
		err  error
		want bool
	}{
		{name: "401 is auth", err: &statusError{StatusCode: http.StatusUnauthorized, Err: errors.New("bad credentials")}, want: true},
		{name: "403 forbidden is auth", err: &statusError{StatusCode: http.StatusForbidden, Err: errors.New("forbidden")}, want: true},
		{name: "403 rate limit is not auth", err: &statusError{StatusCode: http.StatusForbidden, Err: errors.New("API rate limit exceeded")}, want: false},
		{name: "429 is not auth", err: &statusError{StatusCode: http.StatusTooManyRequests, Err: errors.New("rate limit")}, want: false},
		{name: "text unauthorized is auth", err: errors.New("unauthorized"), want: true},
		{name: "text status 403 is auth", err: errors.New("http status 403: resource not accessible"), want: true},
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
