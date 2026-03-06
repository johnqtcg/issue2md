package github

import (
	"errors"
	"net/http"
	"strings"
)

// StatusCode extracts the wrapped HTTP status code when available.
func StatusCode(err error) (int, bool) {
	var stErr *statusError
	if errors.As(err, &stErr) {
		return stErr.StatusCode, true
	}
	return 0, false
}

// IsRateLimitError reports whether an error is a GitHub rate limit failure.
func IsRateLimitError(err error) bool {
	if err == nil {
		return false
	}

	var stErr *statusError
	if errors.As(err, &stErr) {
		if isRateLimitStatusError(stErr) {
			return true
		}
	}

	return looksLikeRateLimitError(err)
}

// IsAuthError reports whether an error is an authentication or authorization failure.
func IsAuthError(err error) bool {
	if err == nil {
		return false
	}
	if IsRateLimitError(err) {
		return false
	}

	if status, ok := StatusCode(err); ok {
		return status == http.StatusUnauthorized || status == http.StatusForbidden
	}

	text := strings.ToLower(err.Error())
	return strings.Contains(text, "unauthorized") ||
		strings.Contains(text, "http status 401") ||
		strings.Contains(text, "bad credentials") ||
		strings.Contains(text, "forbidden")
}
