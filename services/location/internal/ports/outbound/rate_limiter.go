package outbound

import (
	"context"
	"time"
)

type RateLimiter interface {
	Allow(ctx context.Context, key string, ttl time.Duration) (bool, error)
}
