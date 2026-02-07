package usecase

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/daffahilmyf/ride-hailing/services/ride/internal/domain"
	"github.com/daffahilmyf/ride-hailing/services/ride/internal/ports/outbound"
)

type RideService struct {
	Repo outbound.RideRepo
}

type CreateRideCmd struct {
	RiderID    string
	PickupLat  float64
	PickupLng  float64
	DropoffLat float64
	DropoffLng float64
}

func (s *RideService) CreateRide(ctx context.Context, cmd CreateRideCmd) (domain.Ride, error) {
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

	err := s.Repo.Create(ctx, outbound.Ride{
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
}

func (s *RideService) CancelRide(ctx context.Context, id string, reason string) (domain.Ride, error) {
	rideRow, err := s.Repo.Get(ctx, id)
	if err != nil {
		return domain.Ride{}, err
	}

	ride := domain.Ride{
		ID:         rideRow.ID,
		RiderID:    rideRow.RiderID,
		DriverID:   rideRow.DriverID,
		Status:     domain.RideStatus(rideRow.Status),
		PickupLat:  rideRow.PickupLat,
		PickupLng:  rideRow.PickupLng,
		DropoffLat: rideRow.DropoffLat,
		DropoffLng: rideRow.DropoffLng,
	}

	updated, err := ride.Transition(domain.StatusCancelled)
	if err != nil {
		return domain.Ride{}, err
	}
	if err := s.Repo.UpdateStatus(ctx, updated.ID, string(updated.Status), time.Now().UTC()); err != nil {
		return domain.Ride{}, err
	}
	_ = reason
	return updated, nil
}
