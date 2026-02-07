package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisLimiter struct {
	client   *redis.Client
	requests int
	window   time.Duration
	prefix   string
}

type LimiterOption func(*RedisLimiter)

func WithLimiterPrefix(prefix string) LimiterOption {
	return func(l *RedisLimiter) {
		if prefix != "" {
			l.prefix = prefix
		}
	}
}

func WithLimiterWindow(window time.Duration) LimiterOption {
	return func(l *RedisLimiter) {
		if window > 0 {
			l.window = window
		}
	}
}

func WithLimiterRequests(requests int) LimiterOption {
	return func(l *RedisLimiter) {
		if requests > 0 {
			l.requests = requests
		}
	}
}

func NewRedisLimiter(client *redis.Client, opts ...LimiterOption) *RedisLimiter {
	limiter := &RedisLimiter{
		client:   client,
		requests: 100,
		window:   60 * time.Second,
		prefix:   "rl",
	}
	for _, opt := range opts {
		opt(limiter)
	}
	return limiter
}

func (l *RedisLimiter) Allow(key string) (bool, error) {
	ctx := context.Background()
	rkey := fmt.Sprintf("%s:%s", l.prefix, key)

	pipe := l.client.TxPipeline()
	cnt := pipe.Incr(ctx, rkey)
	pipe.Expire(ctx, rkey, l.window)
	_, err := pipe.Exec(ctx)
	if err != nil {
		return false, err
	}
	return cnt.Val() <= int64(l.requests), nil
}
