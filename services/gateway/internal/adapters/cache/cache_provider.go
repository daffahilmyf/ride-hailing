package cache

import (
	"github.com/daffahilmyf/ride-hailing/services/gateway/internal/infra"
	"github.com/daffahilmyf/ride-hailing/services/gateway/internal/ports/outbound"
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
