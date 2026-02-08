package usecase

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"

	"github.com/daffahilmyf/ride-hailing/services/location/internal/domain"
	"github.com/daffahilmyf/ride-hailing/services/location/internal/ports/outbound"
)

type LocationService struct {
	Repo           outbound.LocationRepo
	Publisher      outbound.EventPublisher
	RateLimiter    outbound.RateLimiter
	PublishEnabled bool
	LocationTTL    time.Duration
	MinUpdateGap   time.Duration
	RateKeyPrefix  string
	Clock          Clock
	IDGen          IDGenerator
}

func (s *LocationService) UpdateDriverLocation(ctx context.Context, driverID string, lat float64, lng float64, accuracy float64) (domain.DriverLocation, error) {
	if s.MinUpdateGap > 0 && s.RateLimiter != nil {
		keyPrefix := s.RateKeyPrefix
		if keyPrefix == "" {
			keyPrefix = "driver:location:rate:"
		}
		allowed, err := s.RateLimiter.Allow(ctx, keyPrefix+driverID, s.MinUpdateGap)
		if err != nil {
			return domain.DriverLocation{}, err
		}
		if !allowed {
			return domain.DriverLocation{}, nil
		}
	}
	location, err := domain.NewDriverLocation(driverID, lat, lng, accuracy, s.now())
	if err != nil {
		return domain.DriverLocation{}, err
	}

	err = s.Repo.Upsert(ctx, outbound.Location{
		DriverID:   location.DriverID,
		Lat:        location.Lat,
		Lng:        location.Lng,
		AccuracyM:  location.AccuracyM,
		RecordedAt: location.RecordedAt,
	}, s.LocationTTL)
	if err != nil {
		return domain.DriverLocation{}, err
	}

	if s.PublishEnabled && s.Publisher != nil {
		traceID := getStringFromContext(ctx, "trace_id")
		requestID := getStringFromContext(ctx, "request_id")
		envelope := domain.NewEventEnvelopeWith("driver.location.updated", "location-service", traceID, requestID, map[string]any{
			"driver_id":        location.DriverID,
			"lat":              location.Lat,
			"lng":              location.Lng,
			"accuracy_m":       location.AccuracyM,
			"recorded_at_unix": location.RecordedAt.Unix(),
		}, s.now(), s.newID())
		payload, err := json.Marshal(envelope)
		if err != nil {
			return domain.DriverLocation{}, err
		}
		if err := s.Publisher.Publish(ctx, "driver.location.updated", payload); err != nil {
			return domain.DriverLocation{}, err
		}
	}

	return location, nil
}

func (s *LocationService) GetDriverLocation(ctx context.Context, driverID string) (domain.DriverLocation, error) {
	location, err := s.Repo.Get(ctx, driverID)
	if err != nil {
		return domain.DriverLocation{}, err
	}
	return domain.DriverLocation{
		DriverID:   location.DriverID,
		Lat:        location.Lat,
		Lng:        location.Lng,
		AccuracyM:  location.AccuracyM,
		RecordedAt: location.RecordedAt,
	}, nil
}

func (s *LocationService) ListNearbyDrivers(ctx context.Context, lat float64, lng float64, radiusMeters float64, limit int) ([]outbound.NearbyDriver, error) {
	return s.Repo.Nearby(ctx, lat, lng, radiusMeters, limit)
}

func getStringFromContext(ctx context.Context, key string) string {
	if ctx == nil {
		return ""
	}
	if val := ctx.Value(key); val != nil {
		if s, ok := val.(string); ok {
			return s
		}
	}
	return ""
}

func (s *LocationService) now() time.Time {
	if s != nil && s.Clock != nil {
		return s.Clock.Now()
	}
	return time.Now().UTC()
}

func (s *LocationService) newID() string {
	if s != nil && s.IDGen != nil {
		return s.IDGen()
	}
	return uuid.NewString()
}
