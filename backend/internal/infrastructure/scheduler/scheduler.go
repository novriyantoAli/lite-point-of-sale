// Package scheduler runs the store's background work: the one daily job the
// single-process service has — the automatic backup. It is infrastructure: a
// loop around a use case, not business logic (ADR-0004).
package scheduler

import (
	"context"
	"log/slog"
	"time"
)

// Run calls `run` immediately, then again every `interval`, until ctx is done.
// An error from `run` is logged and the loop keeps going: one failed backup must
// not stop tomorrow's.
func Run(ctx context.Context, run func(context.Context) error, interval time.Duration, logger *slog.Logger) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		if err := run(ctx); err != nil {
			logger.Error("background job failed", "error", err)
		}

		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}
