package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	ridev1 "github.com/daffahilmyf/ride-hailing/proto/ride/v1"
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
	for _, candidate := range candidates {
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
			continue
		}
		_ = s.NotifyOfferSent(ctx, candidate.DriverID, resp.GetOfferId())
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
