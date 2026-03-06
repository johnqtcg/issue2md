package github

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type sleepFunc func(ctx context.Context, d time.Duration) error

type statusError struct {
	Err        error
	StatusCode int
	RetryAfter time.Duration
	ResetAt    time.Time
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

// NewStatusError builds a classified HTTP status error with any retry metadata from headers.
func NewStatusError(statusCode int, err error, header http.Header) error {
	if err == nil {
		err = errors.New(http.StatusText(statusCode))
	}

	stErr := &statusError{
		StatusCode: statusCode,
		Err:        err,
	}
	populateRetryMetadata(stErr, header, time.Now())
	return stErr
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
		if err := sleeper(ctx, retryDelay(err, backoff, time.Now())); err != nil {
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
		if isRateLimitStatusError(stErr) {
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

func retryDelay(err error, fallback time.Duration, now time.Time) time.Duration {
	var stErr *statusError
	if !errors.As(err, &stErr) {
		return fallback
	}
	if stErr.RetryAfter > 0 {
		return stErr.RetryAfter
	}
	if !stErr.ResetAt.IsZero() {
		delay := stErr.ResetAt.Sub(now)
		if delay > 0 {
			return delay
		}
		return 0
	}
	return fallback
}

func isRateLimitStatusError(err *statusError) bool {
	if err == nil {
		return false
	}
	if err.StatusCode == http.StatusTooManyRequests {
		return true
	}
	if err.StatusCode != http.StatusForbidden {
		return false
	}
	if err.RetryAfter > 0 || !err.ResetAt.IsZero() {
		return true
	}
	return looksLikeRateLimitError(err.Err)
}

func populateRetryMetadata(stErr *statusError, header http.Header, now time.Time) {
	if stErr == nil || header == nil {
		return
	}
	if retryAfter, ok := parseRetryAfter(header, now); ok {
		stErr.RetryAfter = retryAfter
	}
	if resetAt, ok := parseRateLimitReset(header); ok {
		stErr.ResetAt = resetAt
	}
}

func parseRetryAfter(header http.Header, now time.Time) (time.Duration, bool) {
	value := strings.TrimSpace(headerValue(header, "Retry-After"))
	if value == "" {
		return 0, false
	}
	if seconds, err := strconv.Atoi(value); err == nil {
		if seconds < 0 {
			return 0, true
		}
		return time.Duration(seconds) * time.Second, true
	}
	when, err := http.ParseTime(value)
	if err != nil {
		return 0, false
	}
	if delay := when.Sub(now); delay > 0 {
		return delay, true
	}
	return 0, true
}

func parseRateLimitReset(header http.Header) (time.Time, bool) {
	value := strings.TrimSpace(headerValue(header, "X-RateLimit-Reset"))
	if value == "" {
		return time.Time{}, false
	}
	seconds, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return time.Time{}, false
	}
	return time.Unix(seconds, 0).UTC(), true
}

func headerValue(header http.Header, key string) string {
	if header == nil {
		return ""
	}
	if value := header.Get(key); value != "" {
		return value
	}
	for name, values := range header {
		if !strings.EqualFold(name, key) || len(values) == 0 {
			continue
		}
		return values[0]
	}
	return ""
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
