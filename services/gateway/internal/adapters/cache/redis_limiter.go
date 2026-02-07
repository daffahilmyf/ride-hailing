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

func (l *RedisLimiter) Allow(key string) (bool, int, time.Time, error) {
	ctx := context.Background()
	rkey := fmt.Sprintf("%s:%s", l.prefix, key)

	pipe := l.client.TxPipeline()
	cnt := pipe.Incr(ctx, rkey)
	ttl := pipe.TTL(ctx, rkey)
	pipe.Expire(ctx, rkey, l.window)
	_, err := pipe.Exec(ctx)
	if err != nil {
		return false, 0, time.Time{}, err
	}

	count := cnt.Val()
	remaining := l.requests - int(count)
	if remaining < 0 {
		remaining = 0
	}

	reset := time.Now().Add(ttl.Val())
	if ttl.Val() < 0 {
		reset = time.Now().Add(l.window)
	}

	return count <= int64(l.requests), remaining, reset, nil
}
