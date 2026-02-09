package github

import (
	"context"
	"errors"
	"net"
	"slices"
	"testing"
	"time"
)

func TestIsRetryableError(t *testing.T) {
	t.Parallel()

	tcs := []struct {
		name string
		err  error
		want bool
	}{
		{name: "status 429", err: &statusError{StatusCode: 429, Err: errors.New("rate limited")}, want: true},
		{name: "status 500", err: &statusError{StatusCode: 500, Err: errors.New("server error")}, want: true},
		{name: "status 403 rate limit", err: &statusError{StatusCode: 403, Err: errors.New("API rate limit exceeded for user")}, want: true},
		{name: "status 403 forbidden scope", err: &statusError{StatusCode: 403, Err: errors.New("forbidden")}, want: false},
		{name: "status 401", err: &statusError{StatusCode: 401, Err: errors.New("unauthorized")}, want: false},
		{name: "status 404", err: &statusError{StatusCode: 404, Err: errors.New("not found")}, want: false},
		{name: "network timeout", err: temporaryNetError{}, want: true},
		{name: "permanent network error", err: permanentNetError{}, want: false},
		{name: "generic error", err: errors.New("boom"), want: false},
	}

	for _, tc := range tcs {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := isRetryableError(tc.err)
			if got != tc.want {
				t.Fatalf("isRetryableError(%v) = %v, want %v", tc.err, got, tc.want)
			}
		})
	}
}

func TestDoWithRetryBackoffSequence(t *testing.T) {
	t.Parallel()

	var sleeps []time.Duration
	sleepFn := func(_ context.Context, d time.Duration) error {
		sleeps = append(sleeps, d)
		return nil
	}

	attempt := 0
	err := doWithRetry(context.Background(), 3, 2*time.Second, sleepFn, func() error {
		attempt++
		if attempt < 4 {
			return &statusError{StatusCode: 500, Err: errors.New("retry")}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("doWithRetry error = %v, want nil", err)
	}

	want := []time.Duration{2 * time.Second, 4 * time.Second, 8 * time.Second}
	if !slices.Equal(sleeps, want) {
		t.Fatalf("sleep sequence = %v, want %v", sleeps, want)
	}
}

type temporaryNetError struct{}

func (temporaryNetError) Error() string   { return "temporary" }
func (temporaryNetError) Timeout() bool   { return true }
func (temporaryNetError) Temporary() bool { return true }

var _ net.Error = temporaryNetError{}

type permanentNetError struct{}

func (permanentNetError) Error() string   { return "permanent" }
func (permanentNetError) Timeout() bool   { return false }
func (permanentNetError) Temporary() bool { return false }

var _ net.Error = permanentNetError{}
