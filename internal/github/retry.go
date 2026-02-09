package github

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"
)

type sleepFunc func(ctx context.Context, d time.Duration) error

type statusError struct {
	StatusCode int
	Err        error
}

func (e *statusError) Error() string {
	if e == nil {
		return ""
	}
	return fmt.Sprintf("http status %d: %v", e.StatusCode, e.Err)
}

func (e *statusError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

func doWithRetry(ctx context.Context, maxRetries int, initialBackoff time.Duration, sleeper sleepFunc, fn func() error) error {
	if sleeper == nil {
		sleeper = sleepContext
	}

	if maxRetries < 0 {
		return fmt.Errorf("invalid maxRetries %d", maxRetries)
	}
	if initialBackoff <= 0 {
		initialBackoff = DefaultInitialBackoff
	}

	var lastErr error
	backoff := initialBackoff
	for attempt := 0; attempt <= maxRetries; attempt++ {
		if err := ctx.Err(); err != nil {
			return fmt.Errorf("retry canceled: %w", err)
		}

		err := fn()
		if err == nil {
			return nil
		}

		lastErr = err
		if attempt == maxRetries || !isRetryableError(err) {
			break
		}
		if err := sleeper(ctx, backoff); err != nil {
			return fmt.Errorf("sleep before retry: %w", err)
		}
		backoff *= 2
	}

	return fmt.Errorf("retry exhausted: %w", lastErr)
}

func isRetryableError(err error) bool {
	if err == nil {
		return false
	}

	var stErr *statusError
	if errors.As(err, &stErr) {
		if stErr.StatusCode == http.StatusTooManyRequests {
			return true
		}
		if stErr.StatusCode == http.StatusForbidden && looksLikeRateLimitError(stErr.Err) {
			return true
		}
		return stErr.StatusCode >= 500 && stErr.StatusCode <= 599
	}

	var netErr net.Error
	if !errors.As(err, &netErr) {
		return false
	}

	if netErr.Timeout() {
		return true
	}

	type temporary interface {
		Temporary() bool
	}
	if temp, ok := any(netErr).(temporary); ok && temp.Temporary() {
		return true
	}

	return false
}

func looksLikeRateLimitError(err error) bool {
	if err == nil {
		return false
	}

	text := strings.ToLower(err.Error())
	return strings.Contains(text, "rate limit")
}

func sleepContext(ctx context.Context, d time.Duration) error {
	timer := time.NewTimer(d)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
