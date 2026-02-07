package redis

import (
	"context"
	"time"

	"github.com/daffahilmyf/ride-hailing/services/matching/internal/ports/outbound"
	"github.com/redis/go-redis/v9"
)

type DriverRepo struct {
	client       *redis.Client
	geoKey       string
	statusKey    string
	availableKey string
	offerPrefix  string
}

func NewDriverRepo(client *redis.Client, geoKey string, statusKey string, availableKey string, offerPrefix string) *DriverRepo {
	if geoKey == "" {
		geoKey = "drivers:geo"
	}
	if statusKey == "" {
		statusKey = "drivers:status"
	}
	if availableKey == "" {
		availableKey = "drivers:available"
	}
	if offerPrefix == "" {
		offerPrefix = "driver:offer:"
	}
	return &DriverRepo{
		client:       client,
		geoKey:       geoKey,
		statusKey:    statusKey,
		availableKey: availableKey,
		offerPrefix:  offerPrefix,
	}
}

func (r *DriverRepo) UpdateStatus(ctx context.Context, driverID string, status string) error {
	if r == nil || r.client == nil {
		return nil
	}
	pipe := r.client.Pipeline()
	pipe.HSet(ctx, r.statusKey, driverID, status)
	if status == "ONLINE_AVAILABLE" {
		pipe.SAdd(ctx, r.availableKey, driverID)
	} else {
		pipe.SRem(ctx, r.availableKey, driverID)
	}
	_, err := pipe.Exec(ctx)
	return err
}

func (r *DriverRepo) MarkOfferSent(ctx context.Context, driverID string, offerID string, ttlSeconds int) error {
	if r == nil || r.client == nil {
		return nil
	}
	key := r.offerPrefix + driverID
	pipe := r.client.Pipeline()
	pipe.Set(ctx, key, offerID, 0)
	if ttlSeconds > 0 {
		pipe.Expire(ctx, key, time.Duration(ttlSeconds)*time.Second)
	}
	pipe.SRem(ctx, r.availableKey, driverID)
	_, err := pipe.Exec(ctx)
	return err
}

func (r *DriverRepo) Nearby(ctx context.Context, lat float64, lng float64, radiusMeters float64, limit int) ([]outbound.Candidate, error) {
	if r == nil || r.client == nil {
		return nil, nil
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
	candidates := make([]outbound.Candidate, 0, len(results))
	for _, item := range results {
		if item.Name == "" {
			continue
		}
		candidates = append(candidates, outbound.Candidate{
			DriverID:  item.Name,
			DistanceM: item.Dist,
		})
	}
	return candidates, nil
}

func (r *DriverRepo) IsAvailable(ctx context.Context, driverID string) (bool, error) {
	if r == nil || r.client == nil {
		return false, nil
	}
	ok, err := r.client.SIsMember(ctx, r.availableKey, driverID).Result()
	if err != nil {
		return false, err
	}
	return ok, nil
}

func (r *DriverRepo) SetLocation(ctx context.Context, driverID string, lat float64, lng float64) error {
	if r == nil || r.client == nil {
		return nil
	}
	_, err := r.client.GeoAdd(ctx, r.geoKey, &redis.GeoLocation{
		Name:      driverID,
		Longitude: lng,
		Latitude:  lat,
	}).Result()
	return err
}
