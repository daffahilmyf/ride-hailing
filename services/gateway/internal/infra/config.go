package infra

type Config struct {
	ServiceName            string
	HTTPAddr               string
	ShutdownTimeoutSeconds int
	Auth                   AuthConfig
	GRPC                   GRPCConfig
}

type AuthConfig struct {
	Enabled   bool
	JWTSecret string
	Issuer    string
	Audience  string
}

type GRPCConfig struct {
	RideAddr       string
	MatchingAddr   string
	LocationAddr   string
	TimeoutSeconds int
	RetryMax       int
	RetryBackoffMs int
}

func DefaultConfig() Config {
	return Config{
		ServiceName:            "api-gateway",
		HTTPAddr:               ":8080",
		ShutdownTimeoutSeconds: 10,
		Auth: AuthConfig{
			Enabled:   false,
			JWTSecret: "",
			Issuer:    "",
			Audience:  "",
		},
		GRPC: GRPCConfig{
			RideAddr:       "localhost:50051",
			MatchingAddr:   "localhost:50052",
			LocationAddr:   "localhost:50053",
			TimeoutSeconds: 2,
			RetryMax:       2,
			RetryBackoffMs: 100,
		},
	}
}
