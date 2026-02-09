package cli

import (
	"bufio"
	"fmt"
	"os"
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
	_ = r

	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("open input file %q: %w", path, err)
	}
	defer func() {
		closeErr := file.Close()
		if err == nil && closeErr != nil {
			err = fmt.Errorf("close input file %q: %w", path, closeErr)
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
		return fmt.Errorf("scan input file %q: %w", path, scanErr)
	}
	return nil
}
