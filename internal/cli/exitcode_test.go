package cli

import (
	"errors"
	"fmt"
	"testing"

	"github.com/johnqtcg/issue2md/internal/config"
	"github.com/johnqtcg/issue2md/internal/parser"
)

func TestResolveExitCode(t *testing.T) {
	t.Parallel()

	tcs := []struct {
		name     string
		err      error
		isBatch  bool
		failed   int
		wantCode int
	}{
		{name: "success", err: nil, isBatch: false, failed: 0, wantCode: ExitOK},
		{name: "validation error", err: config.NewValidationError("url", "bad"), isBatch: false, failed: 0, wantCode: ExitInvalidArguments},
		{name: "conflict error", err: config.NewConflictError("--a", "--b"), isBatch: false, failed: 0, wantCode: ExitInvalidArguments},
		{name: "invalid github url", err: fmt.Errorf("run single URL: %w", fmt.Errorf("parse URL: %w", parser.ErrInvalidGitHubURL)), isBatch: false, failed: 0, wantCode: ExitInvalidArguments},
		{name: "auth error 401", err: errors.New("http status 401: bad credentials"), isBatch: false, failed: 0, wantCode: ExitAuth},
		{name: "auth error 403", err: errors.New("http status 403: resource not accessible"), isBatch: false, failed: 0, wantCode: ExitAuth},
		{name: "rate limit 403 should not be auth", err: errors.New("http status 403: API rate limit exceeded"), isBatch: false, failed: 0, wantCode: ExitRuntime},
		{name: "output conflict", err: ErrOutputConflict, isBatch: false, failed: 0, wantCode: ExitOutputConflict},
		{name: "batch partial", err: nil, isBatch: true, failed: 1, wantCode: ExitPartialSuccess},
		{name: "generic error", err: errors.New("boom"), isBatch: false, failed: 0, wantCode: ExitRuntime},
	}

	for _, tc := range tcs {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := ResolveExitCode(tc.err, tc.isBatch, tc.failed)
			if got != tc.wantCode {
				t.Fatalf("ResolveExitCode = %d, want %d", got, tc.wantCode)
			}
		})
	}
}
