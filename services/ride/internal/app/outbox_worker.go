package app

import (
	"context"
	"time"

	"github.com/daffahilmyf/ride-hailing/services/ride/internal/ports/outbound"
	"go.uber.org/zap"
)

type OutboxWorker struct {
	Repo        outbound.OutboxRepo
	Publisher   outbound.OutboxPublisher
	Logger      *zap.Logger
	BatchSize   int
	MaxAttempts int
	Interval    time.Duration
}

func (w *OutboxWorker) Run(ctx context.Context) {
	if w.Repo == nil || w.Publisher == nil || w.Logger == nil {
		return
	}
	interval := w.Interval
	if interval <= 0 {
		interval = 2 * time.Second
	}
	batch := w.BatchSize
	if batch <= 0 {
		batch = 25
	}
	maxAttempts := w.MaxAttempts
	if maxAttempts <= 0 {
		maxAttempts = 10
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.flush(ctx, batch, maxAttempts)
		}
	}
}

func (w *OutboxWorker) flush(ctx context.Context, batch int, maxAttempts int) {
	messages, err := w.Repo.Claim(ctx, batch, maxAttempts)
	if err != nil {
		w.Logger.Warn("outbox.claim_failed", zap.Error(err))
		return
	}
	for _, msg := range messages {
		if err := w.Publisher.Publish(ctx, msg.Topic, []byte(msg.Payload)); err != nil {
			nextAttempt := time.Now().UTC().Add(backoffDuration(msg.Attempt))
			if msg.Attempt >= maxAttempts {
				_ = w.Repo.MarkFailed(ctx, msg.ID, err.Error(), time.Time{})
				w.Logger.Warn("outbox.dlq",
					zap.String("id", msg.ID),
					zap.String("topic", msg.Topic),
					zap.Int("attempt", msg.Attempt),
					zap.Error(err),
				)
				continue
			}
			_ = w.Repo.MarkFailed(ctx, msg.ID, err.Error(), nextAttempt)
			w.Logger.Warn("outbox.publish_failed",
				zap.String("id", msg.ID),
				zap.String("topic", msg.Topic),
				zap.Int("attempt", msg.Attempt),
				zap.Error(err),
			)
			continue
		}
		if err := w.Repo.MarkSent(ctx, msg.ID); err != nil {
			w.Logger.Warn("outbox.mark_sent_failed", zap.String("id", msg.ID), zap.Error(err))
		}
	}
}

func backoffDuration(attempt int) time.Duration {
	if attempt < 1 {
		attempt = 1
	}
	if attempt > 6 {
		attempt = 6
	}
	return time.Duration(1<<attempt) * time.Second
}
