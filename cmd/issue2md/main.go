package main

import (
	"context"
	"os"

	"github.com/johnqtcg/issue2md/internal/cli"
)

func main() {
	runner := cli.NewApp(cli.AppDeps{})
	os.Exit(runWithRunner(context.Background(), os.Args[1:], runner))
}

func runWithRunner(ctx context.Context, args []string, runner cli.Runner) int {
	return runner.Run(ctx, args)
}
