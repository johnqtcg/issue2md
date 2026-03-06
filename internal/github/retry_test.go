package github

import (
	"context"
	"errors"
	"net"
	"net/http"
	"slices"
	"strconv"
	"testing"
	"time"
)

func TestIsRetryableError(t *testing.T) {
	t.Parallel()

	tcs := []struct {
		err  error
		name string
		want bool
	}{
		{name: "status 429", err: NewStatusError(http.StatusTooManyRequests, errors.New("rate limited"), nil), want: true},
		{name: "status 500", err: NewStatusError(http.StatusInternalServerError, errors.New("server error"), nil), want: true},
		{name: "status 403 retry after", err: NewStatusError(http.StatusForbidden, errors.New("forbidden"), http.Header{"Retry-After": []string{"5"}}), want: true},
		{name: "status 403 forbidden scope", err: NewStatusError(http.StatusForbidden, errors.New("forbidden"), nil), want: false},
		{name: "status 401", err: NewStatusError(http.StatusUnauthorized, errors.New("unauthorized"), nil), want: false},
		{name: "status 404", err: NewStatusError(http.StatusNotFound, errors.New("not found"), nil), want: false},
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

func TestDoWithRetryUsesRetryAfterHeader(t *testing.T) {
	t.Parallel()

	var sleeps []time.Duration
	sleepFn := func(_ context.Context, d time.Duration) error {
		sleeps = append(sleeps, d)
		return nil
	}

	attempt := 0
	err := doWithRetry(context.Background(), 1, 2*time.Second, sleepFn, func() error {
		attempt++
		if attempt == 1 {
			return NewStatusError(http.StatusTooManyRequests, errors.New("rate limited"), http.Header{"Retry-After": []string{"7"}})
		}
		return nil
	})
	if err != nil {
		t.Fatalf("doWithRetry error = %v, want nil", err)
	}

	want := []time.Duration{7 * time.Second}
	if !slices.Equal(sleeps, want) {
		t.Fatalf("sleep sequence = %v, want %v", sleeps, want)
	}
}

func TestRetryDelayUsesRateLimitResetHeader(t *testing.T) {
	t.Parallel()

	now := time.Unix(1_700_000_000, 0).UTC()
	reset := now.Add(11 * time.Second).Unix()
	err := NewStatusError(http.StatusForbidden, errors.New("forbidden"), http.Header{"X-RateLimit-Reset": []string{strconv.FormatInt(reset, 10)}})

	got := retryDelay(err, 2*time.Second, now)
	if got != 11*time.Second {
		t.Fatalf("retryDelay() = %s, want %s", got, 11*time.Second)
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

func (temporaryNetError) Error() string { return "temporary" }

func (temporaryNetError) Timeout() bool { return true }

func (temporaryNetError) Temporary() bool { return true }

var _ net.Error = temporaryNetError{}

type permanentNetError struct{}

func (permanentNetError) Error() string { return "permanent" }

func (permanentNetError) Timeout() bool { return false }

func (permanentNetError) Temporary() bool { return false }

var _ net.Error = permanentNetError{}
