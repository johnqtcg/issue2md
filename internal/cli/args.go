package cli

import "github.com/johnqtcg/issue2md/internal/config"

// Mode identifies the command execution mode.
type Mode string

const (
	// ModeSingle processes one URL from positional args.
	ModeSingle Mode = "single"
	// ModeBatch processes many URLs from --input-file.
	ModeBatch Mode = "batch"
)

// Args contains validated and normalized command mode inputs.
type Args struct {
	Mode Mode
	URL  string
}

// ValidateArgs validates single-vs-batch mode constraints from config.
func ValidateArgs(cfg config.Config) (Args, error) {
	if cfg.InputFile != "" {
		if cfg.OutputPath == "" {
			return Args{}, config.NewValidationError("output", "--output is required when --input-file is set")
		}
		if cfg.Stdout {
			return Args{}, config.NewConflictError("--stdout", "--input-file")
		}
		if len(cfg.Positional) > 0 {
			return Args{}, config.NewValidationError("url", "positional URL is not allowed when --input-file is set")
		}
		return Args{Mode: ModeBatch}, nil
	}

	if len(cfg.Positional) != 1 {
		return Args{}, config.NewValidationError("url", "exactly one URL is required in single mode")
	}

	return Args{
		Mode: ModeSingle,
		URL:  cfg.Positional[0],
	}, nil
}
