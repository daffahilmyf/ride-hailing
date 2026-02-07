package infra

import "github.com/spf13/viper"

func LoadConfig() Config {
	cfg := DefaultConfig()
	cfg.ServiceName = viper.GetString("service.name")
	cfg.GRPCAddr = viper.GetString("grpc.addr")
	cfg.ShutdownTimeoutSeconds = viper.GetInt("shutdown.timeout")
	cfg.RedisAddr = viper.GetString("redis.addr")
	cfg.RedisPassword = viper.GetString("redis.password")
	cfg.RedisDB = viper.GetInt("redis.db")
	cfg.LocationTTLSeconds = viper.GetInt("location.ttl_seconds")
	cfg.LocationKeyPrefix = viper.GetString("location.key_prefix")
	cfg.GeoKey = viper.GetString("location.geo_key")
	cfg.NATSURL = viper.GetString("nats.url")
	cfg.EventsEnabled = viper.GetBool("events.enabled")
	cfg.InternalAuthEnabled = viper.GetBool("internal_auth.enabled")
	cfg.InternalAuthToken = viper.GetString("internal_auth.token")
	return cfg
}
