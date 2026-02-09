package main

import (
	"context"
	"reflect"
	"testing"
)

type fakeRunner struct {
	code    int
	gotArgs []string
}

func (f *fakeRunner) Run(_ context.Context, args []string) int {
	f.gotArgs = append([]string(nil), args...)
	return f.code
}

func TestRunWithRunnerPassThroughArgsAndExitCode(t *testing.T) {
	t.Parallel()

	runner := &fakeRunner{code: 4}
	args := []string{"--input-file", "urls.txt", "--output", "out"}

	got := runWithRunner(context.Background(), args, runner)
	if got != 4 {
		t.Fatalf("runWithRunner = %d, want 4", got)
	}
	if !reflect.DeepEqual(runner.gotArgs, args) {
		t.Fatalf("runner got args = %#v, want %#v", runner.gotArgs, args)
	}
}
