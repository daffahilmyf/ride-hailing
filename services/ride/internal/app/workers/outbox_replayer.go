package workers

import (
	"context"
	"time"

	"github.com/daffahilmyf/ride-hailing/services/ride/internal/ports/outbound"
	"go.uber.org/zap"
)

type OutboxReplayer struct {
	Repo      outbound.OutboxRepo
	Logger    *zap.Logger
	BatchSize int
}

func (r *OutboxReplayer) ReplayFailed(ctx context.Context) {
	if r.Repo == nil || r.Logger == nil {
		return
	}
	batch := r.BatchSize
	if batch <= 0 {
		batch = 50
	}
	messages, err := r.Repo.Claim(ctx, batch, 1)
	if err != nil {
		r.Logger.Warn("outbox.replay_claim_failed", zap.Error(err))
		return
	}
	for _, msg := range messages {
		r.Logger.Info("outbox.replay_queued", zap.String("id", msg.ID), zap.String("topic", msg.Topic))
		_ = r.Repo.MarkFailed(ctx, msg.ID, "replay", time.Now().UTC())
	}
}
