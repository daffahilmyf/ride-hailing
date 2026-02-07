package cache

import (
	"context"
	"time"
)

type NoopCache struct{}

func NewNoopCache() *NoopCache {
	return &NoopCache{}
}

func (c *NoopCache) Get(ctx context.Context, key string) (string, bool, error) {
	return "", false, nil
}

func (c *NoopCache) Set(ctx context.Context, key, value string, ttl time.Duration) error {
	return nil
}

func (c *NoopCache) GetOrSet(ctx context.Context, key string, ttl time.Duration, loader func() (string, error)) (string, error) {
	return loader()
}

func (c *NoopCache) Delete(ctx context.Context, key string) error {
	return nil
}
