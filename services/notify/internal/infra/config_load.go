package infra

import "github.com/spf13/viper"

func LoadConfig() Config {
	cfg := DefaultConfig()
	cfg.ServiceName = viper.GetString("service.name")
	cfg.HTTPAddr = viper.GetString("http.addr")
	cfg.ShutdownTimeoutSeconds = viper.GetInt("shutdown.timeout")
	cfg.NATSURL = viper.GetString("nats.url")
	cfg.NATSSelfHeal = viper.GetBool("nats.self_heal")
	cfg.EventsEnabled = viper.GetBool("events.enabled")
	cfg.RideSubject = viper.GetString("events.ride_subject")
	cfg.DriverSubject = viper.GetString("events.driver_subject")
	cfg.SSEBufferSize = viper.GetInt("sse.buffer_size")
	cfg.SSEKeepaliveSeconds = viper.GetInt("sse.keepalive_seconds")
	cfg.ReplayBufferSize = viper.GetInt("sse.replay_buffer_size")
	cfg.MetricsEnabled = viper.GetBool("observability.metrics_enabled")
	return cfg
}
