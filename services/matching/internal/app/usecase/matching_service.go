package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"math"
	"sort"
	"time"

	ridev1 "github.com/daffahilmyf/ride-hailing/proto/ride/v1"
	"github.com/daffahilmyf/ride-hailing/services/matching/internal/app/metrics"
	"github.com/daffahilmyf/ride-hailing/services/matching/internal/domain"
	"github.com/daffahilmyf/ride-hailing/services/matching/internal/ports/outbound"
	"google.golang.org/grpc/metadata"
)

type MatchingService struct {
	Repo            outbound.DriverRepo
	RideClient      outbound.RideService
	OfferTTLSeconds int
	MatchRadius     float64
	MatchLimit      int
	InternalToken   string
	OfferRetryMax   int
	OfferBackoffMs  int
	OfferMaxBackoff int
	CandidateTTL    int
	ActiveOfferTTL  int
	CooldownSeconds int
	LockTTLSeconds  int
	RadiusStep      float64
	RadiusMax       float64
	MaxOffers       int
	AvgSpeedKmh     float64
	EtaJitterMs     int
	Metrics         *metrics.MatchingMetrics
	Rand            Rand
	Sleep           Sleeper
}

func (s *MatchingService) UpdateDriverStatus(ctx context.Context, driverID string, status string) error {
	if _, err := domain.ParseStatus(status); err != nil {
		return err
	}
	return s.Repo.UpdateStatus(ctx, driverID, status)
}

func (s *MatchingService) FindCandidates(ctx context.Context, lat float64, lng float64, radiusMeters float64, limit int) ([]outbound.Candidate, error) {
	if radiusMeters <= 0 {
		radiusMeters = s.MatchRadius
	}
	if limit <= 0 {
		limit = s.MatchLimit
	}
	step := s.RadiusStep
	if step <= 0 {
		step = radiusMeters
	}
	maxRadius := s.RadiusMax
	if maxRadius <= 0 {
		maxRadius = radiusMeters
	}
	for current := radiusMeters; current <= maxRadius; current += step {
		candidates, err := s.Repo.Nearby(ctx, lat, lng, current, limit)
		if err != nil {
			return nil, err
		}
		available := make([]outbound.Candidate, 0, len(candidates))
		for _, candidate := range candidates {
			ok, err := s.Repo.IsAvailable(ctx, candidate.DriverID)
			if err != nil {
				return nil, err
			}
			if !ok {
				continue
			}
			cooling, err := s.Repo.IsCoolingDown(ctx, candidate.DriverID)
			if err != nil {
				return nil, err
			}
			if cooling {
				continue
			}
			available = append(available, candidate)
		}
		if len(available) > 0 {
			ordered, err := s.rankCandidates(ctx, available)
			if err != nil {
				return nil, err
			}
			return ordered, nil
		}
	}
	return nil, nil
}

func (s *MatchingService) NotifyOfferSent(ctx context.Context, driverID string, offerID string) error {
	return s.Repo.MarkOfferSent(ctx, driverID, offerID, s.OfferTTLSeconds)
}

func (s *MatchingService) HandleRideRequested(ctx context.Context, payload []byte) error {
	var envelope domain.EventEnvelope
	if err := json.Unmarshal(payload, &envelope); err != nil {
		return err
	}

	data, ok := envelope.Payload.(map[string]any)
	if !ok {
		data, ok = envelope.Data.(map[string]any)
	}
	if !ok {
		return errors.New("invalid payload")
	}

	rideID, _ := data["ride_id"].(string)
	pickupLat, okLat := getFloat(data, "pickup_lat")
	pickupLng, okLng := getFloat(data, "pickup_lng")
	if rideID == "" || !okLat || !okLng {
		return nil
	}

	ctx = withTrace(ctx, envelope.TraceID, envelope.RequestID)
	active, ok, err := s.Repo.GetActiveOffer(ctx, rideID)
	if err != nil {
		return err
	}
	if ok && active.OfferID != "" {
		return nil
	}
	hasCandidates, err := s.Repo.HasRideCandidates(ctx, rideID)
	if err != nil {
		return err
	}
	if hasCandidates {
		return nil
	}
	lockTTL := s.LockTTLSeconds
	if lockTTL <= 0 {
		lockTTL = 10
	}
	locked, err := s.Repo.AcquireRideLock(ctx, rideID, lockTTL)
	if err != nil {
		return err
	}
	if !locked {
		return nil
	}
	candidates, err := s.FindCandidates(ctx, pickupLat, pickupLng, s.MatchRadius, s.MatchLimit)
	if err != nil {
		return err
	}
	if len(candidates) == 0 {
		_ = s.Repo.ReleaseRideLock(ctx, rideID)
		if s.Metrics != nil {
			s.Metrics.IncNoCandidates()
		}
		return s.cancelRide(ctx, rideID, "NO_DRIVER")
	}
	if err := s.seedCandidates(ctx, rideID, candidates); err != nil {
		_ = s.Repo.ReleaseRideLock(ctx, rideID)
		return err
	}
	if err := s.sendNextOffer(ctx, rideID, envelope.RequestID); err != nil {
		_ = s.Repo.ReleaseRideLock(ctx, rideID)
		return err
	}
	return nil
}

func (s *MatchingService) HandleDriverLocation(ctx context.Context, payload []byte) error {
	var envelope domain.EventEnvelope
	if err := json.Unmarshal(payload, &envelope); err != nil {
		return err
	}
	data, ok := envelope.Payload.(map[string]any)
	if !ok {
		return errors.New("invalid payload")
	}
	driverID, _ := data["driver_id"].(string)
	lat, _ := data["lat"].(float64)
	lng, _ := data["lng"].(float64)
	if driverID == "" {
		return nil
	}
	return s.Repo.SetLocation(ctx, driverID, lat, lng)
}

func (s *MatchingService) HandleOfferExpired(ctx context.Context, payload []byte) error {
	return s.handleOfferCompletion(ctx, payload)
}

func (s *MatchingService) HandleOfferDeclined(ctx context.Context, payload []byte) error {
	return s.handleOfferCompletion(ctx, payload)
}

func (s *MatchingService) HandleOfferAccepted(ctx context.Context, payload []byte) error {
	var envelope domain.EventEnvelope
	if err := json.Unmarshal(payload, &envelope); err != nil {
		return err
	}
	data, ok := envelope.Payload.(map[string]any)
	if !ok {
		data, ok = envelope.Data.(map[string]any)
	}
	if !ok {
		return errors.New("invalid payload")
	}
	rideID, _ := data["ride_id"].(string)
	if rideID == "" {
		return nil
	}
	ctx = withTrace(ctx, envelope.TraceID, envelope.RequestID)
	_ = s.Repo.ClearRide(ctx, rideID)
	_ = s.Repo.ReleaseRideLock(ctx, rideID)
	return nil
}

func (s *MatchingService) handleOfferCompletion(ctx context.Context, payload []byte) error {
	var envelope domain.EventEnvelope
	if err := json.Unmarshal(payload, &envelope); err != nil {
		return err
	}
	data, ok := envelope.Payload.(map[string]any)
	if !ok {
		data, ok = envelope.Data.(map[string]any)
	}
	if !ok {
		return errors.New("invalid payload")
	}
	rideID, _ := data["ride_id"].(string)
	offerID, _ := data["offer_id"].(string)
	driverID, _ := data["driver_id"].(string)
	if rideID == "" || offerID == "" {
		return nil
	}
	ctx = withTrace(ctx, envelope.TraceID, envelope.RequestID)
	jitter := s.randIntn(200)
	if jitter > 0 {
		s.sleep(time.Duration(jitter) * time.Millisecond)
	}
	if driverID != "" && s.CooldownSeconds > 0 {
		_ = s.Repo.SetCooldown(ctx, driverID, s.CooldownSeconds)
	}
	active, ok, err := s.Repo.GetActiveOffer(ctx, rideID)
	if err != nil {
		return err
	}
	if !ok || active.OfferID != offerID {
		return nil
	}
	_ = s.Repo.ClearActiveOffer(ctx, rideID)
	return s.sendNextOffer(ctx, rideID, envelope.RequestID)
}

func (s *MatchingService) seedCandidates(ctx context.Context, rideID string, candidates []outbound.Candidate) error {
	if s == nil || s.Repo == nil {
		return nil
	}
	driverIDs := make([]string, 0, len(candidates))
	for _, candidate := range candidates {
		if candidate.DriverID == "" {
			continue
		}
		driverIDs = append(driverIDs, candidate.DriverID)
	}
	if len(driverIDs) == 0 {
		return nil
	}
	ttl := s.CandidateTTL
	if ttl <= 0 {
		ttl = s.OfferTTLSeconds * 3
	}
	return s.Repo.StoreRideCandidates(ctx, rideID, driverIDs, ttl)
}

func (s *MatchingService) sendNextOffer(ctx context.Context, rideID string, idempotencyKey string) error {
	if s == nil || s.Repo == nil {
		return nil
	}
	active, ok, err := s.Repo.GetActiveOffer(ctx, rideID)
	if err != nil {
		return err
	}
	if ok && active.OfferID != "" {
		return nil
	}
	if s.LockTTLSeconds > 0 {
		_ = s.Repo.RefreshRideLock(ctx, rideID, s.LockTTLSeconds)
	}
	offerTTL := s.OfferTTLSeconds
	if offerTTL <= 0 {
		offerTTL = 10
	}
	retries := s.OfferRetryMax
	if retries <= 0 {
		retries = 1
	}
	for attempt := 0; attempt < retries; attempt++ {
		driverID, err := s.Repo.PopRideCandidate(ctx, rideID)
		if err != nil {
			return err
		}
		if driverID == "" {
			if s.Metrics != nil {
				s.Metrics.IncNoCandidates()
			}
			_ = s.Repo.ReleaseRideLock(ctx, rideID)
			_ = s.Repo.ClearRide(ctx, rideID)
			return s.cancelRide(ctx, rideID, "NO_DRIVER")
		}
		if s.MaxOffers > 0 {
			count, err := s.Repo.GetOfferCount(ctx, rideID)
			if err != nil {
				return err
			}
			if count >= s.MaxOffers {
				_ = s.Repo.ReleaseRideLock(ctx, rideID)
				_ = s.Repo.ClearRide(ctx, rideID)
				return s.cancelRide(ctx, rideID, "NO_DRIVER")
			}
		}
		exists, err := s.Repo.HasOffer(ctx, driverID)
		if err != nil {
			return err
		}
		if exists {
			if s.Metrics != nil {
				s.Metrics.IncSkipped()
			}
			continue
		}
		req := &ridev1.CreateOfferRequest{
			RideId:          rideID,
			DriverId:        driverID,
			OfferTtlSeconds: int64(offerTTL),
			IdempotencyKey:  idempotencyKey,
		}
		callCtx := withInternalToken(ctx, s.InternalToken)
		resp, err := s.RideClient.CreateOffer(callCtx, req)
		if err != nil {
			if s.Metrics != nil {
				s.Metrics.IncFailed()
			}
			backoff := s.computeBackoff(attempt+1, s.OfferBackoffMs, s.OfferMaxBackoff)
			if backoff > 0 {
				s.sleep(backoff)
			}
			continue
		}
		_ = s.NotifyOfferSent(callCtx, driverID, resp.GetOfferId())
		activeTTL := s.ActiveOfferTTL
		if activeTTL <= 0 {
			activeTTL = offerTTL
		}
		_ = s.Repo.SetActiveOffer(ctx, rideID, resp.GetOfferId(), driverID, activeTTL)
		_ = s.Repo.SetLastOfferAt(ctx, driverID, time.Now().UTC().Unix())
		_, _ = s.Repo.IncrementOfferCount(ctx, rideID, s.CandidateTTL)
		if s.Metrics != nil {
			s.Metrics.IncSent()
		}
		return nil
	}
	return nil
}

func (s *MatchingService) rankCandidates(ctx context.Context, candidates []outbound.Candidate) ([]outbound.Candidate, error) {
	if len(candidates) <= 1 {
		return candidates, nil
	}
	driverIDs := make([]string, 0, len(candidates))
	for _, candidate := range candidates {
		if candidate.DriverID != "" {
			driverIDs = append(driverIDs, candidate.DriverID)
		}
	}
	lastOffers, err := s.Repo.GetLastOfferAt(ctx, driverIDs)
	if err != nil {
		return nil, err
	}
	avgSpeed := s.AvgSpeedKmh
	if avgSpeed <= 0 {
		avgSpeed = 24
	}
	speedMps := avgSpeed * 1000 / 3600
	jitterRange := s.EtaJitterMs
	sort.SliceStable(candidates, func(i, j int) bool {
		left := candidates[i]
		right := candidates[j]
		leftETA := (left.DistanceM / speedMps) * 1000
		rightETA := (right.DistanceM / speedMps) * 1000
		if jitterRange > 0 {
			leftETA += float64(s.randIntn(jitterRange))
			rightETA += float64(s.randIntn(jitterRange))
		}
		if leftETA != rightETA {
			return leftETA < rightETA
		}
		leftLast := lastOffers[left.DriverID]
		rightLast := lastOffers[right.DriverID]
		if leftLast != rightLast {
			return leftLast < rightLast
		}
		return left.DriverID < right.DriverID
	})
	return candidates, nil
}

func (s *MatchingService) cancelRide(ctx context.Context, rideID string, reason string) error {
	if s == nil || s.RideClient == nil {
		return nil
	}
	callCtx := withInternalToken(ctx, s.InternalToken)
	_, err := s.RideClient.CancelRide(callCtx, &ridev1.CancelRideRequest{
		RideId:    rideID,
		Reason:    reason,
		RequestId: rideID + ":" + reason,
	})
	return err
}

func withTrace(ctx context.Context, traceID string, requestID string) context.Context {
	md := metadata.Pairs()
	if traceID != "" {
		md.Append("x-trace-id", traceID)
	}
	if requestID != "" {
		md.Append("x-request-id", requestID)
	}
	return metadata.NewOutgoingContext(ctx, md)
}

func withInternalToken(ctx context.Context, token string) context.Context {
	if token == "" {
		return ctx
	}
	return metadata.AppendToOutgoingContext(ctx, "x-internal-token", token)
}

func (s *MatchingService) computeBackoff(attempt int, baseMs int, maxMs int) time.Duration {
	if attempt <= 0 {
		return 0
	}
	base := time.Duration(baseMs) * time.Millisecond
	if base <= 0 {
		base = 200 * time.Millisecond
	}
	max := time.Duration(maxMs) * time.Millisecond
	if max <= 0 {
		max = 1500 * time.Millisecond
	}
	backoff := float64(base) * math.Pow(2, float64(attempt-1))
	if backoff > float64(max) {
		backoff = float64(max)
	}
	jitter := time.Duration(s.randIntn(100)) * time.Millisecond
	return time.Duration(backoff) + jitter
}

func getFloat(values map[string]any, key string) (float64, bool) {
	raw, ok := values[key]
	if !ok {
		return 0, false
	}
	switch v := raw.(type) {
	case float64:
		return v, true
	case float32:
		return float64(v), true
	case int:
		return float64(v), true
	case int64:
		return float64(v), true
	default:
		return 0, false
	}
}
