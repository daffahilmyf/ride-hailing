package infra

import "github.com/spf13/viper"

func LoadConfig() Config {
	cfg := DefaultConfig()
	cfg.ServiceName = viper.GetString("service.name")
	cfg.HTTPAddr = viper.GetString("http.addr")
	cfg.HTTPReadTimeoutSeconds = viper.GetInt("http.read_timeout_seconds")
	cfg.HTTPWriteTimeoutSeconds = viper.GetInt("http.write_timeout_seconds")
	cfg.GRPCAddr = viper.GetString("grpc.addr")
	cfg.ShutdownTimeoutSeconds = viper.GetInt("shutdown.timeout")
	cfg.PostgresDSN = viper.GetString("postgres.dsn")
	cfg.RedisAddr = viper.GetString("redis.addr")
	cfg.RedisPassword = viper.GetString("redis.password")
	cfg.RedisDB = viper.GetInt("redis.db")
	cfg.Auth.JWTSecret = viper.GetString("auth.jwt_secret")
	cfg.Auth.Issuer = viper.GetString("auth.issuer")
	cfg.Auth.Audience = viper.GetString("auth.audience")
	cfg.Auth.AccessTTLSeconds = viper.GetInt("auth.access_ttl_seconds")
	cfg.Auth.RefreshTTLSeconds = viper.GetInt("auth.refresh_ttl_seconds")
	cfg.InternalAuth.Enabled = viper.GetBool("internal_auth.enabled")
	cfg.InternalAuth.Token = viper.GetString("internal_auth.token")
	cfg.RateLimit.AuthRequests = viper.GetInt("rate_limit.auth_requests")
	cfg.RateLimit.WindowSeconds = viper.GetInt("rate_limit.window_seconds")
	cfg.RateLimit.KeyPrefix = viper.GetString("rate_limit.key_prefix")
	cfg.SessionLimits.Rider = viper.GetInt("session_limits.rider")
	cfg.SessionLimits.Driver = viper.GetInt("session_limits.driver")
	cfg.Observability.MetricsEnabled = viper.GetBool("observability.metrics_enabled")
	cfg.Observability.MetricsAddr = viper.GetString("observability.metrics_addr")
	return cfg
}
