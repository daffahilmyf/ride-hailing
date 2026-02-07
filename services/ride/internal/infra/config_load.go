package infra

import "github.com/spf13/viper"

func LoadConfig() Config {
	cfg := DefaultConfig()
	cfg.ServiceName = viper.GetString("service.name")
	cfg.GRPCAddr = viper.GetString("grpc.addr")
	cfg.ShutdownTimeoutSeconds = viper.GetInt("shutdown.timeout")
	cfg.PostgresDSN = viper.GetString("postgres.dsn")
	cfg.IdempotencyTTLSeconds = viper.GetInt("idempotency.ttl_seconds")
	return cfg
}
