package cli

import (
	"fmt"
	"strings"

	gh "github.com/johnqtcg/issue2md/internal/github"
)

// ItemStatus indicates per-item run outcome.
type ItemStatus string

const (
	// StatusOK indicates a single item succeeded.
	StatusOK ItemStatus = "OK"
	// StatusFailed indicates a single item failed.
	StatusFailed ItemStatus = "FAILED"
)

// ItemResult stores one URL processing result.
type ItemResult struct {
	URL          string
	ResourceType gh.ResourceType
	Status       ItemStatus
	Reason       string
	OutputPath   string
}

// RunSummary stores overall run stats and per-item outcomes.
type RunSummary struct {
	Total     int
	Succeeded int
	Failed    int
	Items     []ItemResult
}

// BuildSummary computes aggregate counters from item results.
func BuildSummary(items []ItemResult) RunSummary {
	out := RunSummary{
		Total: len(items),
		Items: append([]ItemResult(nil), items...),
	}

	for _, item := range items {
		if item.Status == StatusOK {
			out.Succeeded++
		} else {
			out.Failed++
		}
	}
	return out
}

// FormatSummary renders a human-readable summary with failure details.
func FormatSummary(summary RunSummary) string {
	var b strings.Builder

	fmt.Fprintf(&b, "OK total=%d succeeded=%d failed=%d\n", summary.Total, summary.Succeeded, summary.Failed)
	for _, item := range summary.Items {
		if item.Status != StatusFailed {
			continue
		}
		fmt.Fprintf(
			&b,
			"FAILED url=%s type=%s reason=%s\n",
			item.URL,
			item.ResourceType,
			item.Reason,
		)
	}

	return strings.TrimSuffix(b.String(), "\n")
}
