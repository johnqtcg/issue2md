package cli

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestFileInputReaderReadSkipsEmptyLines(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	input := filepath.Join(tmpDir, "urls.txt")
	content := "https://github.com/octo/repo/issues/1\n\n  \nhttps://github.com/octo/repo/pull/2\n"
	if err := os.WriteFile(input, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile error = %v", err)
	}

	reader := NewFileInputReader()
	var got []string
	err := reader.Read(input, func(line string) error {
		got = append(got, line)
		return nil
	})
	if err != nil {
		t.Fatalf("Read error = %v, want nil", err)
	}

	want := []string{
		"https://github.com/octo/repo/issues/1",
		"https://github.com/octo/repo/pull/2",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("lines = %#v, want %#v", got, want)
	}
}

func TestFileInputReaderReadStopsOnCallbackError(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	input := filepath.Join(tmpDir, "urls.txt")
	content := "u1\nu2\nu3\n"
	if err := os.WriteFile(input, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile error = %v", err)
	}

	reader := NewFileInputReader()
	count := 0
	stopErr := errors.New("stop")
	err := reader.Read(input, func(line string) error {
		count++
		if line == "u2" {
			return stopErr
		}
		return nil
	})
	if !errors.Is(err, stopErr) {
		t.Fatalf("Read error = %v, want stop error", err)
	}
	if count != 2 {
		t.Fatalf("callback count = %d, want 2 (streaming stop)", count)
	}
}
