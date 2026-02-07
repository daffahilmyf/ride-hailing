package cache

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisCache struct {
	client *redis.Client
}

func NewRedisCache(client *redis.Client) *RedisCache {
	return &RedisCache{client: client}
}

func (c *RedisCache) Get(ctx context.Context, key string) (string, bool, error) {
	val, err := c.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	return val, true, nil
}

func (c *RedisCache) Set(ctx context.Context, key, value string, ttl time.Duration) error {
	return c.client.Set(ctx, key, value, ttl).Err()
}

func (c *RedisCache) GetOrSet(ctx context.Context, key string, ttl time.Duration, loader func() (string, error)) (string, error) {
	val, ok, err := c.Get(ctx, key)
	if err != nil {
		return "", err
	}
	if ok {
		return val, nil
	}
	val, err = loader()
	if err != nil {
		return "", err
	}
	if err := c.Set(ctx, key, val, ttl); err != nil {
		return "", err
	}
	return val, nil
}

func (c *RedisCache) Delete(ctx context.Context, key string) error {
	return c.client.Del(ctx, key).Err()
}
