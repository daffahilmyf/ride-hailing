package infra

type Config struct {
	ServiceName            string
	GRPCAddr               string
	ShutdownTimeoutSeconds int
	PostgresDSN            string
	IdempotencyTTLSeconds  int
	NATSURL                string
	NATSSelfHeal           bool
	OutboxEnabled          bool
	OutboxIntervalMillis   int
	OutboxBatchSize        int
	OutboxMaxAttempts      int
	OutboxRetentionHours   int
	OfferExpiryEnabled     bool
	OfferExpiryIntervalMs  int
	OfferExpiryBatchSize   int
	InternalAuthEnabled    bool
	InternalAuthToken      string
	UserAddr               string
	UserBreaker            CircuitBreakerConfig
}

type CircuitBreakerConfig struct {
	Enabled         bool
	MaxRequests     uint32
	IntervalSeconds int
	TimeoutSeconds  int
	FailureRatio    float64
	MinRequests     uint32
}

func DefaultConfig() Config {
	return Config{
		ServiceName:            "ride-service",
		GRPCAddr:               ":50051",
		ShutdownTimeoutSeconds: 10,
		PostgresDSN:            "postgres://ride:ride@postgres:5432/rides?sslmode=disable",
		IdempotencyTTLSeconds:  86400,
		NATSURL:                "nats://nats:4222",
		NATSSelfHeal:           true,
		OutboxEnabled:          true,
		OutboxIntervalMillis:   2000,
		OutboxBatchSize:        25,
		OutboxMaxAttempts:      10,
		OutboxRetentionHours:   168,
		OfferExpiryEnabled:     true,
		OfferExpiryIntervalMs:  5000,
		OfferExpiryBatchSize:   50,
		InternalAuthEnabled:    false,
		InternalAuthToken:      "",
		UserAddr:               "user:50054",
		UserBreaker: CircuitBreakerConfig{
			Enabled:         true,
			MaxRequests:     5,
			IntervalSeconds: 30,
			TimeoutSeconds:  10,
			FailureRatio:    0.5,
			MinRequests:     20,
		},
	}
}
