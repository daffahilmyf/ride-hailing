package outbound

import (
	"context"
	"time"
)

type OutboxMessage struct {
	ID      string
	Topic   string
	Payload string
	Attempt int
}

type OutboxRepo interface {
	Enqueue(ctx context.Context, msg OutboxMessage) error
	Claim(ctx context.Context, limit int, maxAttempts int) ([]OutboxMessage, error)
	MarkSent(ctx context.Context, id string) error
	MarkFailed(ctx context.Context, id string, reason string, nextAttemptAt time.Time) error
	DeleteSentBefore(ctx context.Context, cutoff time.Time) (int64, error)
	ResetFailed(ctx context.Context, limit int) (int64, error)
}

type OutboxPublisher interface {
	Publish(ctx context.Context, subject string, payload []byte) error
}
