package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/daffahilmyf/ride-hailing/services/ride/internal/domain"
	"github.com/daffahilmyf/ride-hailing/services/ride/internal/ports/outbound"
)

type fakeRideRepo struct {
	store map[string]outbound.Ride
}

type fakeOutboxRepo struct {
	messages []outbound.OutboxMessage
}

func (f *fakeOutboxRepo) Enqueue(ctx context.Context, msg outbound.OutboxMessage) error {
	f.messages = append(f.messages, msg)
	return nil
}

func (f *fakeOutboxRepo) Claim(ctx context.Context, limit int, maxAttempts int) ([]outbound.OutboxMessage, error) {
	return nil, nil
}

func (f *fakeOutboxRepo) MarkSent(ctx context.Context, id string) error {
	return nil
}

func (f *fakeOutboxRepo) MarkFailed(ctx context.Context, id string, reason string, nextAttemptAt time.Time) error {
	return nil
}

func (f *fakeOutboxRepo) DeleteSentBefore(ctx context.Context, cutoff time.Time) (int64, error) {
	return 0, nil
}

func (f *fakeOutboxRepo) ResetFailed(ctx context.Context, limit int) (int64, error) {
	return 0, nil
}

type fakeOfferRepo struct {
	store map[string]outbound.RideOffer
}

func (f *fakeOfferRepo) Create(ctx context.Context, offer outbound.RideOffer) error {
	if f.store == nil {
		f.store = map[string]outbound.RideOffer{}
	}
	f.store[offer.ID] = offer
	return nil
}

func (f *fakeOfferRepo) Get(ctx context.Context, id string) (outbound.RideOffer, error) {
	if f.store == nil {
		return outbound.RideOffer{}, outbound.ErrNotFound
	}
	offer, ok := f.store[id]
	if !ok {
		return outbound.RideOffer{}, outbound.ErrNotFound
	}
	return offer, nil
}

func (f *fakeOfferRepo) UpdateStatusIfCurrent(ctx context.Context, id string, currentStatus string, nextStatus string) error {
	if f.store == nil {
		return outbound.ErrNotFound
	}
	offer, ok := f.store[id]
	if !ok {
		return outbound.ErrNotFound
	}
	if offer.Status != currentStatus {
		return outbound.ErrConflict
	}
	offer.Status = nextStatus
	f.store[id] = offer
	return nil
}

func (f *fakeOfferRepo) ListExpired(ctx context.Context, cutoff int64, limit int) ([]outbound.RideOffer, error) {
	return nil, nil
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
	outbox := &fakeOutboxRepo{}
	svc := &RideService{Repo: repo, Outbox: outbox, OfferMetrics: &OfferMetrics{}}

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
	if len(outbox.messages) < 2 {
		t.Fatalf("expected outbox messages, got %d", len(outbox.messages))
	}
}

func TestAssignStartComplete(t *testing.T) {
	repo := newFakeRideRepo()
	outbox := &fakeOutboxRepo{}
	svc := &RideService{Repo: repo, Outbox: outbox, OfferMetrics: &OfferMetrics{}}

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

	if err := repo.UpdateStatusIfCurrent(context.Background(), ride.ID, string(domain.StatusMatching), string(domain.StatusOffered), time.Now().UTC()); err != nil {
		t.Fatalf("mark offered error: %v", err)
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
	if len(outbox.messages) < 4 {
		t.Fatalf("expected outbox messages, got %d", len(outbox.messages))
	}
}

func TestCreateOffer(t *testing.T) {
	offers := &fakeOfferRepo{}
	outbox := &fakeOutboxRepo{}
	svc := &RideService{Offers: offers, Outbox: outbox, OfferMetrics: &OfferMetrics{}}

	offer, err := svc.CreateOffer(context.Background(), StartMatchingCmd{
		RideID:   "ride-1",
		DriverID: "driver-1",
		OfferTTL: 5 * time.Second,
	})
	if err != nil {
		t.Fatalf("create offer error: %v", err)
	}
	if offer.ID == "" {
		t.Fatalf("expected offer id")
	}
	if len(offers.store) != 1 {
		t.Fatalf("expected offer stored, got %d", len(offers.store))
	}
	if len(outbox.messages) != 1 {
		t.Fatalf("expected outbox message, got %d", len(outbox.messages))
	}
}
