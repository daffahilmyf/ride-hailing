package infra

type Config struct {
	ServiceName            string
	GRPCAddr               string
	ShutdownTimeoutSeconds int
	PostgresDSN            string
}

func DefaultConfig() Config {
	return Config{
		ServiceName:            "ride-service",
		GRPCAddr:               ":50051",
		ShutdownTimeoutSeconds: 10,
		PostgresDSN:            "postgres://ride:ride@postgres:5432/rides?sslmode=disable",
	}
}
