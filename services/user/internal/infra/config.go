package infra

type Config struct {
	ServiceName             string
	HTTPAddr                string
	HTTPReadTimeoutSeconds  int
	HTTPWriteTimeoutSeconds int
	GRPCAddr                string
	ShutdownTimeoutSeconds  int
	PostgresDSN             string
	RedisAddr               string
	RedisPassword           string
	RedisDB                 int
	Auth                    AuthConfig
	InternalAuth            InternalAuthConfig
	RateLimit               RateLimitConfig
	Observability           ObservabilityConfig
}

type AuthConfig struct {
	JWTSecret         string
	Issuer            string
	Audience          string
	AccessTTLSeconds  int
	RefreshTTLSeconds int
}

type InternalAuthConfig struct {
	Enabled bool
	Token   string
}

type RateLimitConfig struct {
	AuthRequests  int
	WindowSeconds int
	KeyPrefix     string
}

type ObservabilityConfig struct {
	MetricsEnabled bool
	MetricsAddr    string
}

func DefaultConfig() Config {
	return Config{
		ServiceName:             "user-service",
		HTTPAddr:                ":8081",
		HTTPReadTimeoutSeconds:  5,
		HTTPWriteTimeoutSeconds: 5,
		GRPCAddr:                ":50054",
		ShutdownTimeoutSeconds:  10,
		PostgresDSN:             "postgres://ride:ride@localhost:5432/users?sslmode=disable",
		RedisAddr:               "localhost:6379",
		RedisPassword:           "",
		RedisDB:                 0,
		Auth: AuthConfig{
			JWTSecret:         "",
			Issuer:            "ride-hailing",
			Audience:          "ride-hailing-clients",
			AccessTTLSeconds:  1800,
			RefreshTTLSeconds: 2592000,
		},
		InternalAuth: InternalAuthConfig{
			Enabled: false,
			Token:   "",
		},
		RateLimit: RateLimitConfig{
			AuthRequests:  30,
			WindowSeconds: 60,
			KeyPrefix:     "user:rate:",
		},
		Observability: ObservabilityConfig{
			MetricsEnabled: true,
			MetricsAddr:    ":9096",
		},
	}
}
