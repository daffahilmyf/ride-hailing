package redis

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type RateLimiter struct {
	client *redis.Client
}

func NewRateLimiter(client *redis.Client) *RateLimiter {
	return &RateLimiter{client: client}
}

func (r *RateLimiter) Allow(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	if r == nil || r.client == nil {
		return true, nil
	}
	ok, err := r.client.SetNX(ctx, key, "1", ttl).Result()
	if err != nil {
		return false, err
	}
	return ok, nil
}
