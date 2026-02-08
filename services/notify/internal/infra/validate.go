package infra

import "errors"

func ValidateConfig(cfg Config) error {
	if cfg.HTTPAddr == "" {
		return errors.New("http.addr is required")
	}
	if cfg.EventsEnabled {
		if cfg.NATSURL == "" {
			return errors.New("nats.url is required when events enabled")
		}
		if cfg.RideSubject == "" || cfg.DriverSubject == "" {
			return errors.New("events.*_subject is required when events enabled")
		}
	}
	return nil
}
