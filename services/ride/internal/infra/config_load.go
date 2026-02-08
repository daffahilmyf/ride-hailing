package infra

import "github.com/spf13/viper"

func LoadConfig() Config {
	cfg := DefaultConfig()
	cfg.ServiceName = viper.GetString("service.name")
	cfg.GRPCAddr = viper.GetString("grpc.addr")
	cfg.ShutdownTimeoutSeconds = viper.GetInt("shutdown.timeout")
	cfg.PostgresDSN = viper.GetString("postgres.dsn")
	cfg.IdempotencyTTLSeconds = viper.GetInt("idempotency.ttl_seconds")
	cfg.NATSURL = viper.GetString("nats.url")
	cfg.NATSSelfHeal = viper.GetBool("nats.self_heal")
	cfg.OutboxEnabled = viper.GetBool("outbox.enabled")
	cfg.OutboxIntervalMillis = viper.GetInt("outbox.interval_millis")
	cfg.OutboxBatchSize = viper.GetInt("outbox.batch_size")
	cfg.OutboxMaxAttempts = viper.GetInt("outbox.max_attempts")
	cfg.OutboxRetentionHours = viper.GetInt("outbox.retention_hours")
	cfg.OfferExpiryEnabled = viper.GetBool("offer_expiry.enabled")
	cfg.OfferExpiryIntervalMs = viper.GetInt("offer_expiry.interval_millis")
	cfg.OfferExpiryBatchSize = viper.GetInt("offer_expiry.batch_size")
	cfg.InternalAuthEnabled = viper.GetBool("internal_auth.enabled")
	cfg.InternalAuthToken = viper.GetString("internal_auth.token")
	cfg.UserAddr = viper.GetString("grpc.user_addr")
	cfg.UserBreaker.Enabled = viper.GetBool("circuit_breaker.user.enabled")
	cfg.UserBreaker.MaxRequests = uint32(viper.GetInt("circuit_breaker.user.max_requests"))
	cfg.UserBreaker.IntervalSeconds = viper.GetInt("circuit_breaker.user.interval_seconds")
	cfg.UserBreaker.TimeoutSeconds = viper.GetInt("circuit_breaker.user.timeout_seconds")
	cfg.UserBreaker.FailureRatio = viper.GetFloat64("circuit_breaker.user.failure_ratio")
	cfg.UserBreaker.MinRequests = uint32(viper.GetInt("circuit_breaker.user.min_requests"))
	cfg.UserRequestTimeoutSec = viper.GetInt("grpc.user_request_timeout_seconds")
	cfg.UserRetryMax = viper.GetInt("grpc.user_retry_max")
	cfg.UserRetryBackoffMs = viper.GetInt("grpc.user_retry_backoff_ms")
	return cfg
}
