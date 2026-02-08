package infra

type Config struct {
	ServiceName             string
	HTTPAddr                string
	HTTPReadTimeoutSeconds  int
	HTTPWriteTimeoutSeconds int
	GRPCAddr                string
	ShutdownTimeoutSeconds  int
	PostgresDSN             string
	Auth                    AuthConfig
	InternalAuth            InternalAuthConfig
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
		Observability: ObservabilityConfig{
			MetricsEnabled: true,
			MetricsAddr:    ":9096",
		},
	}
}
