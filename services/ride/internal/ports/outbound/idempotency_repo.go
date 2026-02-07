package outbound

import (
	"context"
	"time"
)

type IdempotencyRepo interface {
	Get(ctx context.Context, key string) (string, bool, error)
	Save(ctx context.Context, key string, response string) error
}

type IdempotencyCleanup interface {
	DeleteBefore(ctx context.Context, cutoff time.Time) (int64, error)
}
