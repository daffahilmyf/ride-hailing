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
	Repo        outbound.RideRepo
	Idempotency outbound.IdempotencyRepo
	TxManager   outbound.TxManager
}

type CreateRideCmd struct {
	RiderID        string
	PickupLat      float64
	PickupLng      float64
	DropoffLat     float64
	DropoffLng     float64
	IdempotencyKey string
}

func (s *RideService) CreateRide(ctx context.Context, cmd CreateRideCmd) (domain.Ride, error) {
	return s.withIdempotency(ctx, cmd.IdempotencyKey, func(repo outbound.RideRepo, idem outbound.IdempotencyRepo) (domain.Ride, error) {
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
		return ride, nil
	})
}

func (s *RideService) CancelRide(ctx context.Context, id string, reason string, idempotencyKey string) (domain.Ride, error) {
	return s.withIdempotency(ctx, idempotencyKey, func(repo outbound.RideRepo, _ outbound.IdempotencyRepo) (domain.Ride, error) {
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
		_ = reason
		return updated, nil
	})
}

func (s *RideService) StartMatching(ctx context.Context, rideID string, idempotencyKey string) (domain.Ride, error) {
	return s.withIdempotency(ctx, idempotencyKey, func(repo outbound.RideRepo, _ outbound.IdempotencyRepo) (domain.Ride, error) {
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
		return updated, nil
	})
}

func (s *RideService) AssignDriver(ctx context.Context, rideID, driverID string, idempotencyKey string) (domain.Ride, error) {
	return s.withIdempotency(ctx, idempotencyKey, func(repo outbound.RideRepo, _ outbound.IdempotencyRepo) (domain.Ride, error) {
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
		return next, nil
	})
}

func (s *RideService) StartRide(ctx context.Context, rideID string, idempotencyKey string) (domain.Ride, error) {
	return s.withIdempotency(ctx, idempotencyKey, func(repo outbound.RideRepo, _ outbound.IdempotencyRepo) (domain.Ride, error) {
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
		return updated, nil
	})
}

func (s *RideService) CompleteRide(ctx context.Context, rideID string, idempotencyKey string) (domain.Ride, error) {
	return s.withIdempotency(ctx, idempotencyKey, func(repo outbound.RideRepo, _ outbound.IdempotencyRepo) (domain.Ride, error) {
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
		return updated, nil
	})
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

func (s *RideService) withIdempotency(ctx context.Context, key string, fn func(repo outbound.RideRepo, idem outbound.IdempotencyRepo) (domain.Ride, error)) (domain.Ride, error) {
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
	var tx outbound.Tx
	if s.TxManager != nil {
		var err error
		tx, err = s.TxManager.Begin()
		if err != nil {
			return domain.Ride{}, err
		}
		repo = tx.RideRepo()
		idem = tx.IdempotencyRepo()
		defer func() {
			if tx != nil {
				_ = tx.Rollback()
			}
		}()
	}

	ride, err := fn(repo, idem)
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
