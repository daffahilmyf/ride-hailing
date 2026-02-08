package redis

import (
	"context"
	"strconv"
	"time"

	"github.com/daffahilmyf/ride-hailing/services/location/internal/app/metrics"
	"github.com/daffahilmyf/ride-hailing/services/location/internal/ports/outbound"
	"github.com/redis/go-redis/v9"
)

type LocationRepo struct {
	client     *redis.Client
	keyPrefix  string
	lastPrefix string
	geoKey     string
	metrics    *metrics.LocationMetrics
}

func NewLocationRepo(client *redis.Client, keyPrefix string, geoKey string, metrics *metrics.LocationMetrics) *LocationRepo {
	if keyPrefix == "" {
		keyPrefix = "driver:location:"
	}
	if geoKey == "" {
		geoKey = "drivers:geo"
	}
	return &LocationRepo{
		client:     client,
		keyPrefix:  keyPrefix,
		lastPrefix: keyPrefix + "last:",
		geoKey:     geoKey,
		metrics:    metrics,
	}
}

func (r *LocationRepo) Upsert(ctx context.Context, location outbound.Location, ttl time.Duration) error {
	if r == nil || r.client == nil {
		return nil
	}
	key := r.key(location.DriverID)
	lastKey := r.lastKey(location.DriverID)
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
	if ttl > 0 {
		pipe.Set(ctx, lastKey, location.RecordedAt.Unix(), ttl)
	} else {
		pipe.Set(ctx, lastKey, location.RecordedAt.Unix(), 0)
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

func (r *LocationRepo) Nearby(ctx context.Context, lat float64, lng float64, radiusMeters float64, limit int) ([]outbound.NearbyDriver, error) {
	if r == nil || r.client == nil {
		return nil, outbound.ErrNotFound
	}
	if limit <= 0 {
		limit = 10
	}
	results, err := r.client.GeoRadius(ctx, r.geoKey, lng, lat, &redis.GeoRadiusQuery{
		Radius:    radiusMeters,
		Unit:      "m",
		WithDist:  true,
		WithCoord: true,
		Count:     limit,
		Sort:      "ASC",
	}).Result()
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return []outbound.NearbyDriver{}, nil
	}

	pipe := r.client.Pipeline()
	existsCmds := make([]*redis.IntCmd, len(results))
	for i, item := range results {
		if item.Name == "" {
			continue
		}
		existsCmds[i] = pipe.Exists(ctx, r.lastKey(item.Name))
	}
	if _, err := pipe.Exec(ctx); err != nil {
		return nil, err
	}

	drivers := make([]outbound.NearbyDriver, 0, len(results))
	var stale []string
	for i, item := range results {
		if item.Name == "" {
			continue
		}
		cmd := existsCmds[i]
		if cmd == nil || cmd.Val() == 0 {
			stale = append(stale, item.Name)
			continue
		}
		drivers = append(drivers, outbound.NearbyDriver{
			DriverID:  item.Name,
			Lat:       item.Latitude,
			Lng:       item.Longitude,
			DistanceM: item.Dist,
		})
	}
	if len(stale) > 0 {
		_ = r.client.ZRem(ctx, r.geoKey, stale).Err()
		if r.metrics != nil {
			r.metrics.IncStaleGeoRemoved(len(stale))
		}
	}
	return drivers, nil
}

func (r *LocationRepo) key(driverID string) string {
	return r.keyPrefix + driverID
}

func (r *LocationRepo) lastKey(driverID string) string {
	return r.lastPrefix + driverID
}

func (r *LocationRepo) CleanStaleGeo(ctx context.Context, cursor uint64, count int) (uint64, int, error) {
	if r == nil || r.client == nil {
		return cursor, 0, nil
	}
	if count <= 0 {
		count = 1000
	}
	items, next, err := r.client.ZScan(ctx, r.geoKey, cursor, "", int64(count)).Result()
	if err != nil {
		return cursor, 0, err
	}
	if len(items) == 0 {
		return next, 0, nil
	}
	members := make([]string, 0, len(items)/2)
	for i := 0; i < len(items); i += 2 {
		if items[i] != "" {
			members = append(members, items[i])
		}
	}
	if len(members) == 0 {
		return next, 0, nil
	}
	pipe := r.client.Pipeline()
	existsCmds := make([]*redis.IntCmd, len(members))
	for i, member := range members {
		existsCmds[i] = pipe.Exists(ctx, r.lastKey(member))
	}
	if _, err := pipe.Exec(ctx); err != nil {
		return next, 0, err
	}
	stale := make([]string, 0, len(members))
	for i, member := range members {
		cmd := existsCmds[i]
		if cmd == nil || cmd.Val() == 0 {
			stale = append(stale, member)
		}
	}
	if len(stale) == 0 {
		return next, 0, nil
	}
	if err := r.client.ZRem(ctx, r.geoKey, stale).Err(); err != nil {
		return next, 0, err
	}
	if r.metrics != nil {
		r.metrics.IncStaleGeoRemoved(len(stale))
	}
	return next, len(stale), nil
}

func parseFloat(val string) (float64, error) {
	return strconv.ParseFloat(val, 64)
}

func parseInt(val string) (int64, error) {
	return strconv.ParseInt(val, 10, 64)
}
