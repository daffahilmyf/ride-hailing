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
	cfg.OutboxEnabled = viper.GetBool("outbox.enabled")
	cfg.OutboxIntervalMillis = viper.GetInt("outbox.interval_millis")
	cfg.OutboxBatchSize = viper.GetInt("outbox.batch_size")
	cfg.OutboxMaxAttempts = viper.GetInt("outbox.max_attempts")
	cfg.OutboxRetentionHours = viper.GetInt("outbox.retention_hours")
	cfg.OfferExpiryEnabled = viper.GetBool("offer_expiry.enabled")
	cfg.OfferExpiryIntervalMs = viper.GetInt("offer_expiry.interval_millis")
	cfg.OfferExpiryBatchSize = viper.GetInt("offer_expiry.batch_size")
	return cfg
}
