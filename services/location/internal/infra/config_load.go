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
	cfg.RateLimitEnabled = viper.GetBool("rate_limit.enabled")
	cfg.RateLimitMinGapMs = viper.GetInt("rate_limit.min_gap_ms")
	cfg.RateLimitKeyPrefix = viper.GetString("rate_limit.key_prefix")
	cfg.NATSURL = viper.GetString("nats.url")
	cfg.NATSSelfHeal = viper.GetBool("nats.self_heal")
	cfg.EventsEnabled = viper.GetBool("events.enabled")
	cfg.InternalAuthEnabled = viper.GetBool("internal_auth.enabled")
	cfg.InternalAuthToken = viper.GetString("internal_auth.token")
	cfg.Observability.MetricsEnabled = viper.GetBool("observability.metrics_enabled")
	cfg.Observability.MetricsAddr = viper.GetString("observability.metrics_addr")
	cfg.Observability.TracingEnabled = viper.GetBool("observability.tracing_enabled")
	cfg.Observability.TracingEndpoint = viper.GetString("observability.tracing_endpoint")
	cfg.Observability.TracingInsecure = viper.GetBool("observability.tracing_insecure")
	return cfg
}
