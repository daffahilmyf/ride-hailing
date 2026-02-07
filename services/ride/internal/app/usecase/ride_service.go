package usecase

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"

	"github.com/daffahilmyf/ride-hailing/services/ride/internal/domain"
	"github.com/daffahilmyf/ride-hailing/services/ride/internal/ports/outbound"
)

type RideService struct {
	Repo         outbound.RideRepo
	Idempotency  outbound.IdempotencyRepo
	TxManager    outbound.TxManager
	Outbox       outbound.OutboxRepo
	Offers       outbound.RideOfferRepo
	OfferMetrics *OfferMetrics
}

type CreateRideCmd struct {
	RiderID        string
	PickupLat      float64
	PickupLng      float64
	DropoffLat     float64
	DropoffLng     float64
	IdempotencyKey string
}

type StartMatchingCmd struct {
	RideID         string
	DriverID       string
	OfferTTL       time.Duration
	IdempotencyKey string
}

type OfferActionCmd struct {
	OfferID        string
	IdempotencyKey string
}

func (s *RideService) CreateRide(ctx context.Context, cmd CreateRideCmd) (domain.Ride, error) {
	return s.withIdempotency(ctx, cmd.IdempotencyKey, func(repo outbound.RideRepo, idem outbound.IdempotencyRepo, outbox outbound.OutboxRepo) (domain.Ride, error) {
		now := time.Now().UTC()
		ride := domain.Ride{
			ID:         uuid.NewString(),
			RiderID:    cmd.RiderID,
			Status:     domain.StatusRequested,
			PickupLat:  cmd.PickupLat,
			PickupLng:  cmd.PickupLng,
			DropoffLat: cmd.DropoffLat,
			DropoffLng: cmd.DropoffLng,
		}

		err := repo.Create(ctx, outbound.Ride{
			ID:         ride.ID,
			RiderID:    ride.RiderID,
			DriverID:   ride.DriverID,
			Status:     string(ride.Status),
			PickupLat:  ride.PickupLat,
			PickupLng:  ride.PickupLng,
			DropoffLat: ride.DropoffLat,
			DropoffLng: ride.DropoffLng,
			CreatedAt:  now,
			UpdatedAt:  now,
		})
		if err != nil {
			return domain.Ride{}, err
		}
		if err := enqueueEvent(ctx, outbox, "ride.requested", map[string]any{
			"ride_id":    ride.ID,
			"rider_id":   ride.RiderID,
			"pickup_lat": ride.PickupLat,
			"pickup_lng": ride.PickupLng,
		}); err != nil {
			return domain.Ride{}, err
		}
		return ride, nil
	})
}

func (s *RideService) CancelRide(ctx context.Context, id string, reason string, idempotencyKey string) (domain.Ride, error) {
	return s.withIdempotency(ctx, idempotencyKey, func(repo outbound.RideRepo, _ outbound.IdempotencyRepo, outbox outbound.OutboxRepo) (domain.Ride, error) {
		ride, err := s.loadRide(ctx, id, repo)
		if err != nil {
			return domain.Ride{}, err
		}

		updated, err := ride.Transition(domain.StatusCancelled)
		if err != nil {
			return domain.Ride{}, err
		}
		if err := repo.UpdateStatusIfCurrent(ctx, updated.ID, string(ride.Status), string(updated.Status), time.Now().UTC()); err != nil {
			return domain.Ride{}, err
		}
		if err := enqueueEvent(ctx, outbox, "ride.cancelled", map[string]string{
			"ride_id":  updated.ID,
			"reason":   reason,
			"status":   string(updated.Status),
			"rider_id": updated.RiderID,
		}); err != nil {
			return domain.Ride{}, err
		}
		_ = reason
		return updated, nil
	})
}

func (s *RideService) StartMatching(ctx context.Context, rideID string, idempotencyKey string) (domain.Ride, error) {
	return s.withIdempotency(ctx, idempotencyKey, func(repo outbound.RideRepo, _ outbound.IdempotencyRepo, outbox outbound.OutboxRepo) (domain.Ride, error) {
		ride, err := s.loadRide(ctx, rideID, repo)
		if err != nil {
			return domain.Ride{}, err
		}
		updated, err := ride.Transition(domain.StatusMatching)
		if err != nil {
			return domain.Ride{}, err
		}
		if err := repo.UpdateStatusIfCurrent(ctx, updated.ID, string(ride.Status), string(updated.Status), time.Now().UTC()); err != nil {
			return domain.Ride{}, err
		}
		if err := enqueueEvent(ctx, outbox, "ride.matching.started", map[string]string{
			"ride_id": updated.ID,
			"status":  string(updated.Status),
		}); err != nil {
			return domain.Ride{}, err
		}
		return updated, nil
	})
}

func (s *RideService) AssignDriver(ctx context.Context, rideID, driverID string, idempotencyKey string) (domain.Ride, error) {
	return s.withIdempotency(ctx, idempotencyKey, func(repo outbound.RideRepo, _ outbound.IdempotencyRepo, outbox outbound.OutboxRepo) (domain.Ride, error) {
		ride, err := s.loadRide(ctx, rideID, repo)
		if err != nil {
			return domain.Ride{}, err
		}

		next, err := ride.Transition(domain.StatusDriverAssigned)
		if err != nil {
			return domain.Ride{}, err
		}
		next.DriverID = &driverID

		if err := repo.AssignDriverIfCurrent(ctx, next.ID, driverID, string(ride.Status), string(next.Status), time.Now().UTC()); err != nil {
			return domain.Ride{}, err
		}
		if err := enqueueEvent(ctx, outbox, "ride.driver.assigned", map[string]string{
			"ride_id":   next.ID,
			"driver_id": driverID,
			"status":    string(next.Status),
		}); err != nil {
			return domain.Ride{}, err
		}
		return next, nil
	})
}

func (s *RideService) StartRide(ctx context.Context, rideID string, idempotencyKey string) (domain.Ride, error) {
	return s.withIdempotency(ctx, idempotencyKey, func(repo outbound.RideRepo, _ outbound.IdempotencyRepo, outbox outbound.OutboxRepo) (domain.Ride, error) {
		ride, err := s.loadRide(ctx, rideID, repo)
		if err != nil {
			return domain.Ride{}, err
		}
		updated, err := ride.Transition(domain.StatusInProgress)
		if err != nil {
			return domain.Ride{}, err
		}
		if err := repo.UpdateStatusIfCurrent(ctx, updated.ID, string(ride.Status), string(updated.Status), time.Now().UTC()); err != nil {
			return domain.Ride{}, err
		}
		if err := enqueueEvent(ctx, outbox, "ride.in_progress", map[string]string{
			"ride_id": updated.ID,
			"status":  string(updated.Status),
		}); err != nil {
			return domain.Ride{}, err
		}
		return updated, nil
	})
}

func (s *RideService) CompleteRide(ctx context.Context, rideID string, idempotencyKey string) (domain.Ride, error) {
	return s.withIdempotency(ctx, idempotencyKey, func(repo outbound.RideRepo, _ outbound.IdempotencyRepo, outbox outbound.OutboxRepo) (domain.Ride, error) {
		ride, err := s.loadRide(ctx, rideID, repo)
		if err != nil {
			return domain.Ride{}, err
		}
		updated, err := ride.Transition(domain.StatusCompleted)
		if err != nil {
			return domain.Ride{}, err
		}
		if err := repo.UpdateStatusIfCurrent(ctx, updated.ID, string(ride.Status), string(updated.Status), time.Now().UTC()); err != nil {
			return domain.Ride{}, err
		}
		if err := enqueueEvent(ctx, outbox, "ride.completed", map[string]string{
			"ride_id": updated.ID,
			"status":  string(updated.Status),
		}); err != nil {
			return domain.Ride{}, err
		}
		return updated, nil
	})
}

func (s *RideService) CreateOffer(ctx context.Context, cmd StartMatchingCmd) (domain.RideOffer, error) {
	return s.withIdempotencyOffer(ctx, cmd.IdempotencyKey, func(offers outbound.RideOfferRepo, idem outbound.IdempotencyRepo, outbox outbound.OutboxRepo) (domain.RideOffer, error) {
		ttl := cmd.OfferTTL
		if ttl <= 0 {
			ttl = 15 * time.Second
		}
		offer := domain.NewRideOffer(cmd.RideID, cmd.DriverID, ttl)
		if err := offers.Create(ctx, outbound.RideOffer{
			ID:        offer.ID,
			RideID:    offer.RideID,
			DriverID:  offer.DriverID,
			Status:    string(offer.Status),
			ExpiresAt: offer.ExpiresAt.Unix(),
			CreatedAt: offer.CreatedAt.Unix(),
		}); err != nil {
			return domain.RideOffer{}, err
		}
		s.OfferMetrics.IncCreated()
		if err := enqueueEvent(ctx, outbox, "ride.offer.sent", map[string]string{
			"ride_id":   offer.RideID,
			"driver_id": offer.DriverID,
			"offer_id":  offer.ID,
			"status":    string(offer.Status),
		}); err != nil {
			return domain.RideOffer{}, err
		}
		return offer, nil
	})
}

func (s *RideService) AcceptOffer(ctx context.Context, cmd OfferActionCmd) (domain.RideOffer, error) {
	return s.updateOffer(ctx, cmd, domain.OfferAccepted, "ride.offer.accepted")
}

func (s *RideService) DeclineOffer(ctx context.Context, cmd OfferActionCmd) (domain.RideOffer, error) {
	return s.updateOffer(ctx, cmd, domain.OfferDeclined, "ride.offer.declined")
}

func (s *RideService) ExpireOffer(ctx context.Context, cmd OfferActionCmd) (domain.RideOffer, error) {
	return s.updateOffer(ctx, cmd, domain.OfferExpired, "ride.offer.expired")
}

func (s *RideService) loadRide(ctx context.Context, id string, repo outbound.RideRepo) (domain.Ride, error) {
	rideRow, err := repo.Get(ctx, id)
	if err != nil {
		return domain.Ride{}, err
	}

	return domain.Ride{
		ID:         rideRow.ID,
		RiderID:    rideRow.RiderID,
		DriverID:   rideRow.DriverID,
		Status:     domain.RideStatus(rideRow.Status),
		PickupLat:  rideRow.PickupLat,
		PickupLng:  rideRow.PickupLng,
		DropoffLat: rideRow.DropoffLat,
		DropoffLng: rideRow.DropoffLng,
	}, nil
}

func (s *RideService) withIdempotency(ctx context.Context, key string, fn func(repo outbound.RideRepo, idem outbound.IdempotencyRepo, outbox outbound.OutboxRepo) (domain.Ride, error)) (domain.Ride, error) {
	if key != "" && s.Idempotency != nil {
		if val, ok, err := s.Idempotency.Get(ctx, key); err == nil && ok {
			var ride domain.Ride
			if err := json.Unmarshal([]byte(val), &ride); err == nil {
				return ride, nil
			}
		}
	}

	repo := s.Repo
	idem := s.Idempotency
	outbox := s.Outbox
	var tx outbound.Tx
	if s.TxManager != nil {
		var err error
		tx, err = s.TxManager.Begin()
		if err != nil {
			return domain.Ride{}, err
		}
		repo = tx.RideRepo()
		idem = tx.IdempotencyRepo()
		outbox = tx.OutboxRepo()
		defer func() {
			if tx != nil {
				_ = tx.Rollback()
			}
		}()
	}

	ride, err := fn(repo, idem, outbox)
	if err != nil {
		return domain.Ride{}, err
	}

	if key != "" && idem != nil {
		if b, err := json.Marshal(ride); err == nil {
			_ = idem.Save(ctx, key, string(b))
		}
	}

	if tx != nil {
		if err := tx.Commit(); err != nil {
			return domain.Ride{}, err
		}
	}
	return ride, nil
}

func (s *RideService) withIdempotencyOffer(ctx context.Context, key string, fn func(offers outbound.RideOfferRepo, idem outbound.IdempotencyRepo, outbox outbound.OutboxRepo) (domain.RideOffer, error)) (domain.RideOffer, error) {
	if key != "" && s.Idempotency != nil {
		if val, ok, err := s.Idempotency.Get(ctx, key); err == nil && ok {
			var offer domain.RideOffer
			if err := json.Unmarshal([]byte(val), &offer); err == nil {
				return offer, nil
			}
		}
	}

	offers := s.Offers
	idem := s.Idempotency
	outbox := s.Outbox
	var tx outbound.Tx
	if s.TxManager != nil {
		var err error
		tx, err = s.TxManager.Begin()
		if err != nil {
			return domain.RideOffer{}, err
		}
		offers = tx.RideOfferRepo()
		idem = tx.IdempotencyRepo()
		outbox = tx.OutboxRepo()
		defer func() {
			if tx != nil {
				_ = tx.Rollback()
			}
		}()
	}

	offer, err := fn(offers, idem, outbox)
	if err != nil {
		return domain.RideOffer{}, err
	}

	if key != "" && idem != nil {
		if b, err := json.Marshal(offer); err == nil {
			_ = idem.Save(ctx, key, string(b))
		}
	}

	if tx != nil {
		if err := tx.Commit(); err != nil {
			return domain.RideOffer{}, err
		}
	}
	return offer, nil
}

func enqueueEvent(ctx context.Context, outbox outbound.OutboxRepo, topic string, payload any) error {
	if outbox == nil {
		return nil
	}
	traceID := getStringFromContext(ctx, "trace_id")
	requestID := getStringFromContext(ctx, "request_id")
	envelope := domain.NewEventEnvelope(topic, "ride-service", traceID, requestID, payload)
	event, err := domain.NewOutboxEvent(topic, envelope)
	if err != nil {
		return err
	}
	return outbox.Enqueue(ctx, outbound.OutboxMessage{
		ID:      event.ID,
		Topic:   event.Topic,
		Payload: string(event.Payload),
	})
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

func (s *RideService) updateOffer(ctx context.Context, cmd OfferActionCmd, next domain.RideOfferStatus, topic string) (domain.RideOffer, error) {
	return s.withIdempotencyOffer(ctx, cmd.IdempotencyKey, func(offers outbound.RideOfferRepo, _ outbound.IdempotencyRepo, outbox outbound.OutboxRepo) (domain.RideOffer, error) {
		row, err := offers.Get(ctx, cmd.OfferID)
		if err != nil {
			return domain.RideOffer{}, err
		}
		offer := domain.RideOffer{
			ID:        row.ID,
			RideID:    row.RideID,
			DriverID:  row.DriverID,
			Status:    domain.RideOfferStatus(row.Status),
			ExpiresAt: time.Unix(row.ExpiresAt, 0).UTC(),
			CreatedAt: time.Unix(row.CreatedAt, 0).UTC(),
		}

		updated, err := offer.Transition(next)
		if err != nil {
			return domain.RideOffer{}, err
		}
		if err := offers.UpdateStatusIfCurrent(ctx, updated.ID, string(offer.Status), string(updated.Status)); err != nil {
			return domain.RideOffer{}, err
		}
		switch next {
		case domain.OfferAccepted:
			s.OfferMetrics.IncAccepted()
		case domain.OfferDeclined:
			s.OfferMetrics.IncDeclined()
		case domain.OfferExpired:
			s.OfferMetrics.IncExpired()
		}
		if err := enqueueEvent(ctx, outbox, topic, map[string]string{
			"offer_id":  updated.ID,
			"ride_id":   updated.RideID,
			"driver_id": updated.DriverID,
			"status":    string(updated.Status),
		}); err != nil {
			return domain.RideOffer{}, err
		}
		return updated, nil
	})
}
