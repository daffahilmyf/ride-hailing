package infra

type Config struct {
	ServiceName            string
	HTTPAddr               string
	ShutdownTimeoutSeconds int
	NATSURL                string
	NATSSelfHeal           bool
	EventsEnabled          bool
	RideSubject            string
	DriverSubject          string
	SSEBufferSize          int
	SSEKeepaliveSeconds    int
	ReplayBufferSize       int
	MetricsEnabled         bool
}

func DefaultConfig() Config {
	return Config{
		ServiceName:            "notify-service",
		HTTPAddr:               ":8090",
		ShutdownTimeoutSeconds: 10,
		NATSURL:                "nats://nats:4222",
		NATSSelfHeal:           true,
		EventsEnabled:          true,
		RideSubject:            "ride.>",
		DriverSubject:          "driver.>",
		SSEBufferSize:          64,
		SSEKeepaliveSeconds:    15,
		ReplayBufferSize:       256,
		MetricsEnabled:         true,
	}
}
