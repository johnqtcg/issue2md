package config

import (
	"errors"
	"strings"
	"testing"
)

func TestValidationError(t *testing.T) {
	t.Parallel()

	err := NewValidationError("format", "must be markdown")

	var vErr *ValidationError
	if !errors.As(err, &vErr) {
		t.Fatalf("error type = %T, want *ValidationError", err)
	}
	if vErr.Field != "format" {
		t.Fatalf("Field = %q, want format", vErr.Field)
	}
	if !strings.Contains(err.Error(), "must be markdown") {
		t.Fatalf("error message = %q, want contains %q", err.Error(), "must be markdown")
	}
}

func TestConflictError(t *testing.T) {
	t.Parallel()

	err := NewConflictError("--stdout", "--input-file")

	var cErr *ConflictError
	if !errors.As(err, &cErr) {
		t.Fatalf("error type = %T, want *ConflictError", err)
	}
	if cErr.Left != "--stdout" || cErr.Right != "--input-file" {
		t.Fatalf("conflict = (%q,%q), want (%q,%q)", cErr.Left, cErr.Right, "--stdout", "--input-file")
	}
	if !strings.Contains(err.Error(), "--stdout") || !strings.Contains(err.Error(), "--input-file") {
		t.Fatalf("error message = %q, want options in message", err.Error())
	}
}

func TestWrapError(t *testing.T) {
	t.Parallel()

	base := errors.New("boom")
	err := WrapError("load flags", base)
	if err == nil {
		t.Fatal("WrapError returned nil")
	}
	if !errors.Is(err, base) {
		t.Fatalf("wrapped error does not contain base error: %v", err)
	}
	if !strings.Contains(err.Error(), "load flags") {
		t.Fatalf("wrapped message = %q, want contains %q", err.Error(), "load flags")
	}
}
