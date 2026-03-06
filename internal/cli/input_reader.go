package cli

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// InputReader streams URLs from an input file.
type InputReader interface {
	Read(path string, handle func(line string) error) error
}

type fileInputReader struct{}

// NewFileInputReader creates a streaming line-by-line input reader.
func NewFileInputReader() InputReader {
	return &fileInputReader{}
}

func (r *fileInputReader) Read(path string, handle func(line string) error) (err error) {
	cleanPath := filepath.Clean(path)
	if cleanPath == "." {
		return fmt.Errorf("open input file %q: path is empty", path)
	}

	// #nosec G304 -- this CLI intentionally opens a user-specified local file path.
	file, err := os.Open(cleanPath)
	if err != nil {
		return fmt.Errorf("open input file %q: %w", cleanPath, err)
	}
	defer func() {
		closeErr := file.Close()
		if err == nil && closeErr != nil {
			err = fmt.Errorf("close input file %q: %w", cleanPath, closeErr)
		}
	}()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		if handleErr := handle(line); handleErr != nil {
			return fmt.Errorf("process input line %q: %w", line, handleErr)
		}
	}
	if scanErr := scanner.Err(); scanErr != nil {
		return fmt.Errorf("scan input file %q: %w", cleanPath, scanErr)
	}
	return nil
}
