package cache

import (
	"github.com/daffahilmyf/ride-hailing/services/gateway/internal/infra"
	"github.com/daffahilmyf/ride-hailing/services/gateway/internal/ports/outbound"
	"time"
)

func NewCache(cfg infra.CacheConfig, redisCfg infra.RedisConfig) outbound.Cache {
	if !cfg.Enabled {
		return NewNoopCache()
	}
	client := NewRedisClient(RedisConfig{
		Addr:     redisCfg.Addr,
		Password: redisCfg.Password,
		DB:       redisCfg.DB,
	})
	return NewRedisCache(client)
}

func DefaultTTL(cfg infra.CacheConfig) time.Duration {
	if cfg.DefaultTTLSeconds <= 0 {
		return 60 * time.Second
	}
	return time.Duration(cfg.DefaultTTLSeconds) * time.Second
}
