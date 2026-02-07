package outbound

import (
	"context"
	"time"
)

type Cache interface {
	Get(ctx context.Context, key string) (string, bool, error)
	Set(ctx context.Context, key, value string, ttl time.Duration) error
	GetOrSet(ctx context.Context, key string, ttl time.Duration, loader func() (string, error)) (string, error)
	Delete(ctx context.Context, key string) error
}
