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

func NewRedisLimiter(client *redis.Client, requests int, window time.Duration, prefix string) *RedisLimiter {
	return &RedisLimiter{
		client:   client,
		requests: requests,
		window:   window,
		prefix:   prefix,
	}
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
