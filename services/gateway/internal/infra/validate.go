package infra

import "errors"

func ValidateConfig(cfg Config) error {
	if cfg.HTTPAddr == "" {
		return errors.New("http.addr is required")
	}
	if cfg.Auth.Enabled && cfg.Auth.JWTSecret == "" {
		return errors.New("auth.jwt_secret is required when auth enabled")
	}
	if cfg.GRPC.RideAddr == "" || cfg.GRPC.MatchingAddr == "" || cfg.GRPC.LocationAddr == "" || cfg.GRPC.UserAddr == "" {
		return errors.New("grpc.*_addr is required")
	}
	if cfg.RateLimit.Requests <= 0 || cfg.RateLimit.WindowSeconds <= 0 {
		return errors.New("rate_limit.requests and rate_limit.window_seconds must be > 0")
	}
	if cfg.Redis.Addr == "" {
		return errors.New("redis.addr is required")
	}
	if cfg.Observability.TracingEnabled && cfg.Observability.TracingEndpoint == "" {
		return errors.New("observability.tracing_endpoint is required when tracing enabled")
	}
	return nil
}
