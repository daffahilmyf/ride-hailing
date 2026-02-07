package handlers

import (
	"context"
	"time"

	"github.com/daffahilmyf/ride-hailing/services/gateway/internal/ports/outbound"
)

type ReadinessCache struct {
	Cache outbound.Cache
	TTL   time.Duration
	Key   string
}

func (r ReadinessCache) Get(ctx context.Context) (string, bool, error) {
	if r.Cache == nil {
		return "", false, nil
	}
	return r.Cache.Get(ctx, r.Key)
}

func (r ReadinessCache) Set(ctx context.Context, value string) error {
	if r.Cache == nil {
		return nil
	}
	return r.Cache.Set(ctx, r.Key, value, r.TTL)
}
