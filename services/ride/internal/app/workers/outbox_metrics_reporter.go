package workers

import (
	"context"
	"time"

	"github.com/daffahilmyf/ride-hailing/services/ride/internal/app/metrics"
	"go.uber.org/zap"
)

type MetricsReporter struct {
	Outbox   *metrics.OutboxMetrics
	Logger   *zap.Logger
	Interval time.Duration
}

func (r *MetricsReporter) Run(ctx context.Context) {
	if r.Outbox == nil || r.Logger == nil {
		return
	}
	interval := r.Interval
	if interval <= 0 {
		interval = 30 * time.Second
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			r.Logger.Info("outbox.metrics",
				zap.Int64("claimed", r.Outbox.Claimed()),
				zap.Int64("published", r.Outbox.Published()),
				zap.Int64("failed", r.Outbox.Failed()),
				zap.Int64("dlq", r.Outbox.DLQ()),
			)
		}
	}
}
