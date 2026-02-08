package redisadapter

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

func (r *RateLimiter) Allow(ctx context.Context, key string, limit int, window time.Duration) (bool, error) {
	if r == nil || r.client == nil {
		return true, nil
	}
	if key == "" || limit <= 0 || window <= 0 {
		return false, nil
	}
	pipe := r.client.Pipeline()
	countCmd := pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, window)
	if _, err := pipe.Exec(ctx); err != nil {
		return false, err
	}
	return countCmd.Val() <= int64(limit), nil
}
