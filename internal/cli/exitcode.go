package cli

import (
	"errors"

	"github.com/johnqtcg/issue2md/internal/config"
	gh "github.com/johnqtcg/issue2md/internal/github"
	"github.com/johnqtcg/issue2md/internal/parser"
)

const (
	// ExitOK indicates all items completed successfully.
	ExitOK = 0
	// ExitRuntime indicates generic runtime failure.
	ExitRuntime = 1
	// ExitInvalidArguments indicates invalid CLI arguments.
	ExitInvalidArguments = 2
	// ExitAuth indicates auth/authz failures.
	ExitAuth = 3
	// ExitPartialSuccess indicates at least one failure in batch mode.
	ExitPartialSuccess = 4
	// ExitOutputConflict indicates output file conflict without force mode.
	ExitOutputConflict = 5
)

// ResolveExitCode maps run error state to CLI exit codes.
func ResolveExitCode(err error, isBatch bool, failed int) int {
	if isBatch && failed > 0 {
		return ExitPartialSuccess
	}
	if err == nil {
		return ExitOK
	}

	var vErr *config.ValidationError
	if errors.As(err, &vErr) {
		return ExitInvalidArguments
	}

	var cErr *config.ConflictError
	if errors.As(err, &cErr) {
		return ExitInvalidArguments
	}
	if errors.Is(err, parser.ErrInvalidGitHubURL) {
		return ExitInvalidArguments
	}

	if errors.Is(err, ErrOutputConflict) {
		return ExitOutputConflict
	}

	if gh.IsAuthError(err) {
		return ExitAuth
	}

	return ExitRuntime
}
