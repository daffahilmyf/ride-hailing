package infra

type Config struct {
	ServiceName            string
	GRPCAddr               string
	ShutdownTimeoutSeconds int
	RedisAddr              string
	RedisPassword          string
	RedisDB                int
	GeoKey                 string
	StatusKey              string
	AvailableKey           string
	OfferKeyPrefix         string
	OfferTTLSeconds        int
	MatchRadiusMeters      float64
	MatchLimit             int
	OfferRetryMax          int
	OfferRetryBackoffMs    int
	OfferRetryMaxBackoffMs int
	NATSURL                string
	EventsEnabled          bool
	RideRequestedSubject   string
	DriverLocationSubject  string
	InternalAuthEnabled    bool
	InternalAuthToken      string
	RideServiceAddr        string
	RideServiceToken       string
	Observability          ObservabilityConfig
}

type ObservabilityConfig struct {
	MetricsEnabled  bool
	MetricsAddr     string
	TracingEnabled  bool
	TracingEndpoint string
	TracingInsecure bool
}

func DefaultConfig() Config {
	return Config{
		ServiceName:            "matching-service",
		GRPCAddr:               ":50052",
		ShutdownTimeoutSeconds: 10,
		RedisAddr:              "redis:6379",
		RedisPassword:          "",
		RedisDB:                0,
		GeoKey:                 "drivers:geo",
		StatusKey:              "drivers:status",
		AvailableKey:           "drivers:available",
		OfferKeyPrefix:         "driver:offer:",
		OfferTTLSeconds:        10,
		MatchRadiusMeters:      3000,
		MatchLimit:             5,
		OfferRetryMax:          3,
		OfferRetryBackoffMs:    200,
		OfferRetryMaxBackoffMs: 1500,
		NATSURL:                "nats://nats:4222",
		EventsEnabled:          true,
		RideRequestedSubject:   "ride.requested",
		DriverLocationSubject:  "driver.location.updated",
		InternalAuthEnabled:    false,
		InternalAuthToken:      "",
		RideServiceAddr:        "ride:50051",
		RideServiceToken:       "",
		Observability: ObservabilityConfig{
			MetricsEnabled:  true,
			MetricsAddr:     ":9096",
			TracingEnabled:  false,
			TracingEndpoint: "",
			TracingInsecure: true,
		},
	}
}
