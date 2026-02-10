package redis

import (
	"context"
	"strconv"
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
	ridePrefix   string
	activePrefix string
	cooldownKey  string
	lockPrefix   string
	offerCount   string
	lastOfferKey string
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
	ridePrefix := "ride:candidates:"
	activePrefix := "ride:active_offer:"
	cooldownKey := "driver:cooldown"
	lockPrefix := "ride:lock:"
	offerCount := "ride:offer_count:"
	lastOfferKey := "driver:last_offer"
	return &DriverRepo{
		client:       client,
		geoKey:       geoKey,
		statusKey:    statusKey,
		availableKey: availableKey,
		offerPrefix:  offerPrefix,
		ridePrefix:   ridePrefix,
		activePrefix: activePrefix,
		cooldownKey:  cooldownKey,
		lockPrefix:   lockPrefix,
		offerCount:   offerCount,
		lastOfferKey: lastOfferKey,
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

func (r *DriverRepo) HasOffer(ctx context.Context, driverID string) (bool, error) {
	if r == nil || r.client == nil {
		return false, nil
	}
	key := r.offerPrefix + driverID
	ok, err := r.client.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return ok == 1, nil
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

func (r *DriverRepo) SetCooldown(ctx context.Context, driverID string, ttlSeconds int) error {
	if r == nil || r.client == nil {
		return nil
	}
	if driverID == "" || ttlSeconds <= 0 {
		return nil
	}
	return r.client.Set(ctx, r.cooldownKey+":"+driverID, "1", time.Duration(ttlSeconds)*time.Second).Err()
}

func (r *DriverRepo) IsCoolingDown(ctx context.Context, driverID string) (bool, error) {
	if r == nil || r.client == nil {
		return false, nil
	}
	if driverID == "" {
		return false, nil
	}
	ok, err := r.client.Exists(ctx, r.cooldownKey+":"+driverID).Result()
	if err != nil {
		return false, err
	}
	return ok == 1, nil
}

func (r *DriverRepo) AcquireRideLock(ctx context.Context, rideID string, ttlSeconds int) (bool, error) {
	if r == nil || r.client == nil {
		return true, nil
	}
	if rideID == "" {
		return false, nil
	}
	ttl := time.Duration(ttlSeconds) * time.Second
	if ttl <= 0 {
		ttl = 10 * time.Second
	}
	return r.client.SetNX(ctx, r.lockPrefix+rideID, "1", ttl).Result()
}

func (r *DriverRepo) RefreshRideLock(ctx context.Context, rideID string, ttlSeconds int) error {
	if r == nil || r.client == nil {
		return nil
	}
	if rideID == "" {
		return nil
	}
	ttl := time.Duration(ttlSeconds) * time.Second
	if ttl <= 0 {
		ttl = 10 * time.Second
	}
	return r.client.Expire(ctx, r.lockPrefix+rideID, ttl).Err()
}

func (r *DriverRepo) ReleaseRideLock(ctx context.Context, rideID string) error {
	if r == nil || r.client == nil {
		return nil
	}
	if rideID == "" {
		return nil
	}
	return r.client.Del(ctx, r.lockPrefix+rideID).Err()
}

func (r *DriverRepo) IncrementOfferCount(ctx context.Context, rideID string, ttlSeconds int) (int, error) {
	if r == nil || r.client == nil {
		return 0, nil
	}
	if rideID == "" {
		return 0, nil
	}
	key := r.offerCount + rideID
	pipe := r.client.TxPipeline()
	incr := pipe.Incr(ctx, key)
	if ttlSeconds > 0 {
		pipe.Expire(ctx, key, time.Duration(ttlSeconds)*time.Second)
	}
	if _, err := pipe.Exec(ctx); err != nil {
		return 0, err
	}
	val, err := incr.Result()
	return int(val), err
}

func (r *DriverRepo) GetOfferCount(ctx context.Context, rideID string) (int, error) {
	if r == nil || r.client == nil {
		return 0, nil
	}
	if rideID == "" {
		return 0, nil
	}
	val, err := r.client.Get(ctx, r.offerCount+rideID).Int()
	if err == redis.Nil {
		return 0, nil
	}
	return val, err
}

func (r *DriverRepo) HasRideCandidates(ctx context.Context, rideID string) (bool, error) {
	if r == nil || r.client == nil {
		return false, nil
	}
	if rideID == "" {
		return false, nil
	}
	count, err := r.client.LLen(ctx, r.ridePrefix+rideID).Result()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *DriverRepo) SetLastOfferAt(ctx context.Context, driverID string, tsUnix int64) error {
	if r == nil || r.client == nil {
		return nil
	}
	if driverID == "" {
		return nil
	}
	return r.client.HSet(ctx, r.lastOfferKey, driverID, tsUnix).Err()
}

func (r *DriverRepo) GetLastOfferAt(ctx context.Context, driverIDs []string) (map[string]int64, error) {
	if r == nil || r.client == nil {
		return nil, nil
	}
	if len(driverIDs) == 0 {
		return map[string]int64{}, nil
	}
	fields := make([]string, 0, len(driverIDs))
	for _, id := range driverIDs {
		if id == "" {
			continue
		}
		fields = append(fields, id)
	}
	if len(fields) == 0 {
		return map[string]int64{}, nil
	}
	vals, err := r.client.HMGet(ctx, r.lastOfferKey, fields...).Result()
	if err != nil {
		return nil, err
	}
	out := make(map[string]int64, len(fields))
	for i, field := range fields {
		if i >= len(vals) || vals[i] == nil {
			continue
		}
		switch v := vals[i].(type) {
		case string:
			if parsed, err := strconv.ParseInt(v, 10, 64); err == nil {
				out[field] = parsed
			}
		case int64:
			out[field] = v
		case int:
			out[field] = int64(v)
		}
	}
	return out, nil
}

func (r *DriverRepo) StoreRideCandidates(ctx context.Context, rideID string, driverIDs []string, ttlSeconds int) error {
	if r == nil || r.client == nil {
		return nil
	}
	if rideID == "" || len(driverIDs) == 0 {
		return nil
	}
	key := r.ridePrefix + rideID
	pipe := r.client.TxPipeline()
	pipe.Del(ctx, key)
	values := make([]interface{}, 0, len(driverIDs))
	for _, id := range driverIDs {
		if id == "" {
			continue
		}
		values = append(values, id)
	}
	if len(values) == 0 {
		_, err := pipe.Exec(ctx)
		return err
	}
	pipe.RPush(ctx, key, values...)
	if ttlSeconds > 0 {
		pipe.Expire(ctx, key, time.Duration(ttlSeconds)*time.Second)
	}
	_, err := pipe.Exec(ctx)
	return err
}

func (r *DriverRepo) PopRideCandidate(ctx context.Context, rideID string) (string, error) {
	if r == nil || r.client == nil {
		return "", nil
	}
	key := r.ridePrefix + rideID
	val, err := r.client.LPop(ctx, key).Result()
	if err == redis.Nil {
		return "", nil
	}
	return val, err
}

func (r *DriverRepo) SetActiveOffer(ctx context.Context, rideID string, offerID string, driverID string, ttlSeconds int) error {
	if r == nil || r.client == nil {
		return nil
	}
	if rideID == "" || offerID == "" || driverID == "" {
		return nil
	}
	key := r.activePrefix + rideID
	pipe := r.client.TxPipeline()
	pipe.HSet(ctx, key, "offer_id", offerID, "driver_id", driverID)
	if ttlSeconds > 0 {
		pipe.Expire(ctx, key, time.Duration(ttlSeconds)*time.Second)
	}
	_, err := pipe.Exec(ctx)
	return err
}

func (r *DriverRepo) GetActiveOffer(ctx context.Context, rideID string) (outbound.ActiveOffer, bool, error) {
	if r == nil || r.client == nil {
		return outbound.ActiveOffer{}, false, nil
	}
	key := r.activePrefix + rideID
	vals, err := r.client.HMGet(ctx, key, "offer_id", "driver_id").Result()
	if err != nil {
		return outbound.ActiveOffer{}, false, err
	}
	if len(vals) != 2 || vals[0] == nil || vals[1] == nil {
		return outbound.ActiveOffer{}, false, nil
	}
	offerID, _ := vals[0].(string)
	driverID, _ := vals[1].(string)
	if offerID == "" || driverID == "" {
		return outbound.ActiveOffer{}, false, nil
	}
	return outbound.ActiveOffer{OfferID: offerID, DriverID: driverID}, true, nil
}

func (r *DriverRepo) ClearActiveOffer(ctx context.Context, rideID string) error {
	if r == nil || r.client == nil {
		return nil
	}
	if rideID == "" {
		return nil
	}
	return r.client.Del(ctx, r.activePrefix+rideID).Err()
}

func (r *DriverRepo) ClearRide(ctx context.Context, rideID string) error {
	if r == nil || r.client == nil {
		return nil
	}
	if rideID == "" {
		return nil
	}
	keys := []string{
		r.ridePrefix + rideID,
		r.activePrefix + rideID,
		r.offerCount + rideID,
		r.lockPrefix + rideID,
	}
	return r.client.Del(ctx, keys...).Err()
}
