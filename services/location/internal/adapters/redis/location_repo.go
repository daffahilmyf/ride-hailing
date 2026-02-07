package redis

import (
	"context"
	"strconv"
	"time"

	"github.com/daffahilmyf/ride-hailing/services/location/internal/ports/outbound"
	"github.com/redis/go-redis/v9"
)

type LocationRepo struct {
	client    *redis.Client
	keyPrefix string
	geoKey    string
}

func NewLocationRepo(client *redis.Client, keyPrefix string, geoKey string) *LocationRepo {
	if keyPrefix == "" {
		keyPrefix = "driver:location:"
	}
	if geoKey == "" {
		geoKey = "drivers:geo"
	}
	return &LocationRepo{client: client, keyPrefix: keyPrefix, geoKey: geoKey}
}

func (r *LocationRepo) Upsert(ctx context.Context, location outbound.Location, ttl time.Duration) error {
	if r == nil || r.client == nil {
		return nil
	}
	key := r.key(location.DriverID)
	pipe := r.client.Pipeline()
	pipe.HSet(ctx, key, map[string]any{
		"lat":              location.Lat,
		"lng":              location.Lng,
		"accuracy_m":       location.AccuracyM,
		"recorded_at_unix": location.RecordedAt.Unix(),
	})
	if ttl > 0 {
		pipe.Expire(ctx, key, ttl)
	}
	pipe.GeoAdd(ctx, r.geoKey, &redis.GeoLocation{
		Name:      location.DriverID,
		Longitude: location.Lng,
		Latitude:  location.Lat,
	})
	_, err := pipe.Exec(ctx)
	return err
}

func (r *LocationRepo) Get(ctx context.Context, driverID string) (outbound.Location, error) {
	if r == nil || r.client == nil {
		return outbound.Location{}, outbound.ErrNotFound
	}
	values, err := r.client.HGetAll(ctx, r.key(driverID)).Result()
	if err != nil {
		return outbound.Location{}, err
	}
	if len(values) == 0 {
		return outbound.Location{}, outbound.ErrNotFound
	}
	lat, err := parseFloat(values["lat"])
	if err != nil {
		return outbound.Location{}, err
	}
	lng, err := parseFloat(values["lng"])
	if err != nil {
		return outbound.Location{}, err
	}
	accuracy, err := parseFloat(values["accuracy_m"])
	if err != nil {
		return outbound.Location{}, err
	}
	recorded, err := parseInt(values["recorded_at_unix"])
	if err != nil {
		return outbound.Location{}, err
	}
	return outbound.Location{
		DriverID:   driverID,
		Lat:        lat,
		Lng:        lng,
		AccuracyM:  accuracy,
		RecordedAt: time.Unix(recorded, 0).UTC(),
	}, nil
}

func (r *LocationRepo) key(driverID string) string {
	return r.keyPrefix + driverID
}

func parseFloat(val string) (float64, error) {
	return strconv.ParseFloat(val, 64)
}

func parseInt(val string) (int64, error) {
	return strconv.ParseInt(val, 10, 64)
}
