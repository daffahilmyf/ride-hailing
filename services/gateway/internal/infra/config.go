package infra

type Config struct {
	ServiceName            string
	HTTPAddr               string
	ShutdownTimeoutSeconds int
	Auth                   AuthConfig
	GRPC                   GRPCConfig
	RateLimit              RateLimitConfig
	MaxBodyBytes           int64
	Redis                  RedisConfig
	Cache                  CacheConfig
	Observability          ObservabilityConfig
	HTTP                   HTTPConfig
}

type AuthConfig struct {
	Enabled    bool
	JWTSecret  string
	JWTSecrets []string
	Issuer     string
	Audience   string
}

type GRPCConfig struct {
	RideAddr        string
	MatchingAddr    string
	LocationAddr    string
	TimeoutSeconds  int
	RetryMax        int
	RetryBackoffMs  int
	ConnectRequired bool
	InternalToken   string
}

type RateLimitConfig struct {
	Requests      int
	WindowSeconds int
}

type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

type CacheConfig struct {
	Enabled           bool
	DefaultTTLSeconds int
}

type HTTPConfig struct {
	RequestTimeoutSeconds int
	GzipEnabled           bool
}

func DefaultConfig() Config {
	return Config{
		ServiceName:            "api-gateway",
		HTTPAddr:               ":8080",
		ShutdownTimeoutSeconds: 10,
		Auth: AuthConfig{
			Enabled:    false,
			JWTSecret:  "",
			JWTSecrets: []string{},
			Issuer:     "",
			Audience:   "",
		},
		GRPC: GRPCConfig{
			RideAddr:        "localhost:50051",
			MatchingAddr:    "localhost:50052",
			LocationAddr:    "localhost:50053",
			TimeoutSeconds:  2,
			RetryMax:        2,
			RetryBackoffMs:  100,
			ConnectRequired: true,
			InternalToken:   "",
		},
		RateLimit: RateLimitConfig{
			Requests:      100,
			WindowSeconds: 60,
		},
		MaxBodyBytes: 1_048_576,
		Redis: RedisConfig{
			Addr:     "localhost:6379",
			Password: "",
			DB:       0,
		},
		Cache: CacheConfig{
			Enabled:           true,
			DefaultTTLSeconds: 60,
		},
		Observability: ObservabilityConfig{
			MetricsEnabled:  true,
			TracingEnabled:  false,
			TracingEndpoint: "",
			TracingInsecure: true,
		},
		HTTP: HTTPConfig{
			RequestTimeoutSeconds: 5,
			GzipEnabled:           true,
		},
	}
}
