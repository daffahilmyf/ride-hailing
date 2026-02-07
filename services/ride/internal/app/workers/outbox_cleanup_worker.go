package workers

import (
	"context"
	"time"

	"github.com/daffahilmyf/ride-hailing/services/ride/internal/ports/outbound"
	"go.uber.org/zap"
)

type OutboxCleanupWorker struct {
	Repo      outbound.OutboxRepo
	Logger    *zap.Logger
	Retention time.Duration
	Interval  time.Duration
}

func (w *OutboxCleanupWorker) Run(ctx context.Context) {
	if w.Repo == nil || w.Logger == nil || w.Retention <= 0 {
		return
	}
	interval := w.Interval
	if interval <= 0 {
		interval = 24 * time.Hour
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			cutoff := time.Now().UTC().Add(-w.Retention)
			deleted, err := w.Repo.DeleteSentBefore(ctx, cutoff)
			if err != nil {
				w.Logger.Warn("outbox.cleanup_failed", zap.Error(err))
				continue
			}
			if deleted > 0 {
				w.Logger.Info("outbox.cleanup", zap.Int64("deleted", deleted))
			}
		}
	}
}
