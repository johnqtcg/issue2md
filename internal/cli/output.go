package cli

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/johnqtcg/issue2md/internal/config"
	gh "github.com/johnqtcg/issue2md/internal/github"
)

// ErrOutputConflict indicates the output file already exists and force mode is disabled.
var ErrOutputConflict = errors.New("output file already exists")

const outputPathStdout = "stdout"

// OutputWriter writes rendered markdown to stdout or filesystem.
type OutputWriter interface {
	Write(cfg config.Config, mode Mode, ref gh.ResourceRef, markdown []byte) (string, error)
}

type fileOutputWriter struct {
	stdout io.Writer
}

// NewOutputWriter creates an output writer with the provided stdout sink.
func NewOutputWriter(stdout io.Writer) OutputWriter {
	return &fileOutputWriter{stdout: stdout}
}

func (w *fileOutputWriter) Write(cfg config.Config, mode Mode, ref gh.ResourceRef, markdown []byte) (string, error) {
	if cfg.Stdout {
		if _, err := w.stdout.Write(markdown); err != nil {
			return "", fmt.Errorf("write markdown to stdout: %w", err)
		}
		return outputPathStdout, nil
	}

	targetPath, err := resolveOutputPath(cfg, mode, ref)
	if err != nil {
		return "", fmt.Errorf("resolve output path: %w", err)
	}

	if err := ensureWritable(targetPath, cfg.Force); err != nil {
		return "", fmt.Errorf("validate output path %q: %w", targetPath, err)
	}

	parentDir := filepath.Dir(targetPath)
	if err := os.MkdirAll(parentDir, 0o755); err != nil {
		return "", fmt.Errorf("create output directory %q: %w", parentDir, err)
	}

	if err := os.WriteFile(targetPath, markdown, 0o644); err != nil {
		return "", fmt.Errorf("write output file %q: %w", targetPath, err)
	}
	return targetPath, nil
}

func resolveOutputPath(cfg config.Config, mode Mode, ref gh.ResourceRef) (string, error) {
	defaultName, err := defaultFileName(ref)
	if err != nil {
		return "", fmt.Errorf("build default file name: %w", err)
	}

	if mode == ModeBatch {
		if cfg.OutputPath == "" {
			return "", fmt.Errorf("batch output path is empty")
		}
		return filepath.Join(cfg.OutputPath, defaultName), nil
	}

	if cfg.OutputPath == "" {
		return defaultName, nil
	}

	info, err := os.Stat(cfg.OutputPath)
	switch {
	case err == nil && info.IsDir():
		return filepath.Join(cfg.OutputPath, defaultName), nil
	case err == nil:
		return cfg.OutputPath, nil
	case !errors.Is(err, os.ErrNotExist):
		return "", fmt.Errorf("stat output path %q: %w", cfg.OutputPath, err)
	}

	if strings.EqualFold(filepath.Ext(cfg.OutputPath), ".md") {
		return cfg.OutputPath, nil
	}
	return filepath.Join(cfg.OutputPath, defaultName), nil
}

func defaultFileName(ref gh.ResourceRef) (string, error) {
	var resourcePart string
	switch ref.Type {
	case gh.ResourceIssue:
		resourcePart = "issue"
	case gh.ResourcePullRequest:
		resourcePart = "pr"
	case gh.ResourceDiscussion:
		resourcePart = "discussion"
	default:
		return "", fmt.Errorf("unsupported resource type %q", ref.Type)
	}

	return fmt.Sprintf("%s-%s-%s-%d.md", ref.Owner, ref.Repo, resourcePart, ref.Number), nil
}

func ensureWritable(path string, force bool) error {
	_, err := os.Stat(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("stat output file: %w", err)
	}
	if !force {
		return ErrOutputConflict
	}
	return nil
}
