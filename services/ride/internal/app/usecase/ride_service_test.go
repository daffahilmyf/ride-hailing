package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/daffahilmyf/ride-hailing/services/ride/internal/ports/outbound"
)

type fakeRideRepo struct {
	store map[string]outbound.Ride
}

func newFakeRideRepo() *fakeRideRepo {
	return &fakeRideRepo{store: map[string]outbound.Ride{}}
}

func (f *fakeRideRepo) Create(ctx context.Context, ride outbound.Ride) error {
	f.store[ride.ID] = ride
	return nil
}

func (f *fakeRideRepo) Get(ctx context.Context, id string) (outbound.Ride, error) {
	ride, ok := f.store[id]
	if !ok {
		return outbound.Ride{}, outbound.ErrNotFound
	}
	return ride, nil
}

func (f *fakeRideRepo) UpdateStatusIfCurrent(ctx context.Context, id string, currentStatus string, nextStatus string, updatedAt time.Time) error {
	r, ok := f.store[id]
	if !ok {
		return outbound.ErrNotFound
	}
	if r.Status != currentStatus {
		return outbound.ErrConflict
	}
	r.Status = nextStatus
	r.UpdatedAt = updatedAt
	f.store[id] = r
	return nil
}

func (f *fakeRideRepo) AssignDriverIfCurrent(ctx context.Context, id string, driverID string, currentStatus string, nextStatus string, updatedAt time.Time) error {
	r, ok := f.store[id]
	if !ok {
		return outbound.ErrNotFound
	}
	if r.Status != currentStatus {
		return outbound.ErrConflict
	}
	r.DriverID = &driverID
	r.Status = nextStatus
	r.UpdatedAt = updatedAt
	f.store[id] = r
	return nil
}

func TestCreateAndCancelRide(t *testing.T) {
	repo := newFakeRideRepo()
	svc := &RideService{Repo: repo}

	ride, err := svc.CreateRide(context.Background(), CreateRideCmd{
		RiderID:    "r1",
		PickupLat:  1,
		PickupLng:  2,
		DropoffLat: 3,
		DropoffLng: 4,
	})
	if err != nil {
		t.Fatalf("create error: %v", err)
	}

	_, err = svc.CancelRide(context.Background(), ride.ID, "test", "")
	if err != nil {
		t.Fatalf("cancel error: %v", err)
	}
}

func TestAssignStartComplete(t *testing.T) {
	repo := newFakeRideRepo()
	svc := &RideService{Repo: repo}

	ride, err := svc.CreateRide(context.Background(), CreateRideCmd{
		RiderID:    "r1",
		PickupLat:  1,
		PickupLng:  2,
		DropoffLat: 3,
		DropoffLng: 4,
	})
	if err != nil {
		t.Fatalf("create error: %v", err)
	}

	_, err = svc.StartMatching(context.Background(), ride.ID, "")
	if err != nil {
		t.Fatalf("start matching error: %v", err)
	}

	_, err = svc.AssignDriver(context.Background(), ride.ID, "d1", "")
	if err != nil {
		t.Fatalf("assign error: %v", err)
	}

	_, err = svc.StartRide(context.Background(), ride.ID, "")
	if err != nil {
		t.Fatalf("start ride error: %v", err)
	}

	_, err = svc.CompleteRide(context.Background(), ride.ID, "")
	if err != nil {
		t.Fatalf("complete error: %v", err)
	}
}
