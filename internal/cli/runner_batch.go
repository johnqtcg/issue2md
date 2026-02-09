package cli

import (
	"context"
	"fmt"

	"github.com/johnqtcg/issue2md/internal/config"
	"github.com/johnqtcg/issue2md/internal/converter"
	gh "github.com/johnqtcg/issue2md/internal/github"
)

func (a *App) runBatch(ctx context.Context, cfg config.Config, fetcher gh.Fetcher, renderer converter.Renderer) (RunSummary, error) {
	var items []ItemResult

	err := a.inputReader.Read(cfg.InputFile, func(line string) error {
		item, processErr := a.processOne(ctx, cfg, ModeBatch, line, fetcher, renderer)
		if processErr != nil {
			item.Status = StatusFailed
			item.Reason = processErr.Error()
			items = append(items, item)
			writeStatusLine(a.stdout, item)
			return nil
		}

		items = append(items, item)
		writeStatusLine(a.stdout, item)
		return nil
	})
	if err != nil {
		return BuildSummary(items), fmt.Errorf("read batch input file %q: %w", cfg.InputFile, err)
	}

	return BuildSummary(items), nil
}
