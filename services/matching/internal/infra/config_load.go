package infra

import "github.com/spf13/viper"

func LoadConfig() Config {
	cfg := DefaultConfig()
	cfg.ServiceName = viper.GetString("service.name")
	cfg.GRPCAddr = viper.GetString("grpc.addr")
	cfg.ShutdownTimeoutSeconds = viper.GetInt("shutdown.timeout")
	cfg.RedisAddr = viper.GetString("redis.addr")
	cfg.RedisPassword = viper.GetString("redis.password")
	cfg.RedisDB = viper.GetInt("redis.db")
	cfg.GeoKey = viper.GetString("matching.geo_key")
	cfg.StatusKey = viper.GetString("matching.status_key")
	cfg.AvailableKey = viper.GetString("matching.available_key")
	cfg.OfferKeyPrefix = viper.GetString("matching.offer_key_prefix")
	cfg.OfferTTLSeconds = viper.GetInt("matching.offer_ttl_seconds")
	cfg.MatchRadiusMeters = viper.GetFloat64("matching.radius_meters")
	cfg.MatchLimit = viper.GetInt("matching.limit")
	cfg.OfferRetryMax = viper.GetInt("matching.offer_retry_max")
	cfg.OfferRetryBackoffMs = viper.GetInt("matching.offer_retry_backoff_ms")
	cfg.OfferRetryMaxBackoffMs = viper.GetInt("matching.offer_retry_max_backoff_ms")
	cfg.NATSURL = viper.GetString("nats.url")
	cfg.EventsEnabled = viper.GetBool("events.enabled")
	cfg.RideRequestedSubject = viper.GetString("events.ride_requested_subject")
	cfg.DriverLocationSubject = viper.GetString("events.driver_location_subject")
	cfg.InternalAuthEnabled = viper.GetBool("internal_auth.enabled")
	cfg.InternalAuthToken = viper.GetString("internal_auth.token")
	cfg.RideServiceAddr = viper.GetString("ride.addr")
	cfg.RideServiceToken = viper.GetString("ride.internal_token")
	cfg.Observability.MetricsEnabled = viper.GetBool("observability.metrics_enabled")
	cfg.Observability.MetricsAddr = viper.GetString("observability.metrics_addr")
	cfg.Observability.TracingEnabled = viper.GetBool("observability.tracing_enabled")
	cfg.Observability.TracingEndpoint = viper.GetString("observability.tracing_endpoint")
	cfg.Observability.TracingInsecure = viper.GetBool("observability.tracing_insecure")
	return cfg
}
