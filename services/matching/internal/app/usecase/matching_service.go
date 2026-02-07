package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"math"
	"math/rand"
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
	Metrics         *metrics.MatchingMetrics
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
	candidates, err := s.Repo.Nearby(ctx, lat, lng, radiusMeters, limit)
	if err != nil {
		return nil, err
	}
	available := make([]outbound.Candidate, 0, len(candidates))
	for _, candidate := range candidates {
		ok, err := s.Repo.IsAvailable(ctx, candidate.DriverID)
		if err != nil {
			return nil, err
		}
		if ok {
			available = append(available, candidate)
		}
	}
	return available, nil
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
		return errors.New("invalid payload")
	}

	rideID, _ := data["ride_id"].(string)
	pickupLat, okLat := getFloat(data, "pickup_lat")
	pickupLng, okLng := getFloat(data, "pickup_lng")
	if rideID == "" || !okLat || !okLng {
		return nil
	}

	ctx = withTrace(ctx, envelope.TraceID, envelope.RequestID)
	candidates, err := s.FindCandidates(ctx, pickupLat, pickupLng, s.MatchRadius, s.MatchLimit)
	if err != nil {
		return err
	}
	if len(candidates) == 0 {
		if s.Metrics != nil {
			s.Metrics.IncNoCandidates()
		}
		return nil
	}
	retries := s.OfferRetryMax
	if retries <= 0 {
		retries = 1
	}
	attempts := 0
	for _, candidate := range candidates {
		if attempts >= retries {
			break
		}
		exists, err := s.Repo.HasOffer(ctx, candidate.DriverID)
		if err != nil {
			return err
		}
		if exists {
			if s.Metrics != nil {
				s.Metrics.IncSkipped()
			}
			continue
		}
		offerTTL := s.OfferTTLSeconds
		if offerTTL <= 0 {
			offerTTL = 10
		}
		req := &ridev1.CreateOfferRequest{
			RideId:          rideID,
			DriverId:        candidate.DriverID,
			OfferTtlSeconds: int32(offerTTL),
			IdempotencyKey:  envelope.RequestID,
		}
		ctx = withInternalToken(ctx, s.InternalToken)
		resp, err := s.RideClient.CreateOffer(ctx, req)
		if err != nil {
			if s.Metrics != nil {
				s.Metrics.IncFailed()
			}
			attempts++
			backoff := computeBackoff(attempts, s.OfferBackoffMs, s.OfferMaxBackoff)
			if backoff > 0 {
				time.Sleep(backoff)
			}
			continue
		}
		_ = s.NotifyOfferSent(ctx, candidate.DriverID, resp.GetOfferId())
		if s.Metrics != nil {
			s.Metrics.IncSent()
		}
		attempts++
		break
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

func computeBackoff(attempt int, baseMs int, maxMs int) time.Duration {
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
	jitter := time.Duration(rand.Intn(100)) * time.Millisecond
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
