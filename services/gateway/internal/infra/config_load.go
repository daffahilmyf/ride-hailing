package infra

import "github.com/spf13/viper"

func LoadConfig() Config {
	cfg := DefaultConfig()

	cfg.ServiceName = viper.GetString("service.name")
	cfg.HTTPAddr = viper.GetString("http.addr")
	cfg.ShutdownTimeoutSeconds = viper.GetInt("shutdown.timeout")

	cfg.Auth.Enabled = viper.GetBool("auth.enabled")
	cfg.Auth.JWTSecret = viper.GetString("auth.jwt_secret")
	cfg.Auth.Issuer = viper.GetString("auth.issuer")
	cfg.Auth.Audience = viper.GetString("auth.audience")

	cfg.GRPC.RideAddr = viper.GetString("grpc.ride_addr")
	cfg.GRPC.MatchingAddr = viper.GetString("grpc.matching_addr")
	cfg.GRPC.LocationAddr = viper.GetString("grpc.location_addr")
	cfg.GRPC.TimeoutSeconds = viper.GetInt("grpc.timeout_seconds")
	cfg.GRPC.RetryMax = viper.GetInt("grpc.retry_max")
	cfg.GRPC.RetryBackoffMs = viper.GetInt("grpc.retry_backoff_ms")
	cfg.RateLimit.Requests = viper.GetInt("rate_limit.requests")
	cfg.RateLimit.WindowSeconds = viper.GetInt("rate_limit.window_seconds")
	cfg.MaxBodyBytes = viper.GetInt64("http.max_body_bytes")
	cfg.Redis.Addr = viper.GetString("redis.addr")
	cfg.Redis.Password = viper.GetString("redis.password")
	cfg.Redis.DB = viper.GetInt("redis.db")

	return cfg
}
