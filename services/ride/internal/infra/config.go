package infra

type Config struct {
	ServiceName            string
	GRPCAddr               string
	ShutdownTimeoutSeconds int
	PostgresDSN            string
	IdempotencyTTLSeconds  int
	NATSURL                string
	OutboxEnabled          bool
	OutboxIntervalMillis   int
	OutboxBatchSize        int
	OutboxMaxAttempts      int
}

func DefaultConfig() Config {
	return Config{
		ServiceName:            "ride-service",
		GRPCAddr:               ":50051",
		ShutdownTimeoutSeconds: 10,
		PostgresDSN:            "postgres://ride:ride@postgres:5432/rides?sslmode=disable",
		IdempotencyTTLSeconds:  86400,
		NATSURL:                "nats://nats:4222",
		OutboxEnabled:          true,
		OutboxIntervalMillis:   2000,
		OutboxBatchSize:        25,
		OutboxMaxAttempts:      10,
	}
}
