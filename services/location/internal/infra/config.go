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
	NATSURL                string
	EventsEnabled          bool
	InternalAuthEnabled    bool
	InternalAuthToken      string
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
		NATSURL:                "nats://nats:4222",
		EventsEnabled:          true,
		InternalAuthEnabled:    false,
		InternalAuthToken:      "",
	}
}
