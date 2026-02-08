package infra

type Config struct {
	ServiceName            string
	GRPCAddr               string
	ShutdownTimeoutSeconds int
	RedisAddr              string
	RedisPassword          string
	RedisDB                int
	LocationTTLSeconds     int
	LocationKeyPrefix      string
	GeoKey                 string
	RateLimitEnabled       bool
	RateLimitMinGapMs      int
	RateLimitKeyPrefix     string
	NATSURL                string
	NATSSelfHeal           bool
	EventsEnabled          bool
	InternalAuthEnabled    bool
	InternalAuthToken      string
	Observability          ObservabilityConfig
}

func DefaultConfig() Config {
	return Config{
		ServiceName:            "location-service",
		GRPCAddr:               ":50053",
		ShutdownTimeoutSeconds: 10,
		RedisAddr:              "redis:6379",
		RedisPassword:          "",
		RedisDB:                0,
		LocationTTLSeconds:     60,
		LocationKeyPrefix:      "driver:location:",
		GeoKey:                 "drivers:geo",
		RateLimitEnabled:       true,
		RateLimitMinGapMs:      300,
		RateLimitKeyPrefix:     "driver:location:rate:",
		NATSURL:                "nats://nats:4222",
		NATSSelfHeal:           true,
		EventsEnabled:          true,
		InternalAuthEnabled:    false,
		InternalAuthToken:      "",
		Observability: ObservabilityConfig{
			MetricsEnabled:  true,
			MetricsAddr:     ":9095",
			TracingEnabled:  false,
			TracingEndpoint: "",
			TracingInsecure: true,
		},
	}
}

type ObservabilityConfig struct {
	MetricsEnabled  bool
	MetricsAddr     string
	TracingEnabled  bool
	TracingEndpoint string
	TracingInsecure bool
}
