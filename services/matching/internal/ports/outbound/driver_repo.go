package outbound

import "context"

type Candidate struct {
	DriverID  string
	DistanceM float64
}

type DriverRepo interface {
	UpdateStatus(ctx context.Context, driverID string, status string) error
	MarkOfferSent(ctx context.Context, driverID string, offerID string, ttlSeconds int) error
	HasOffer(ctx context.Context, driverID string) (bool, error)
	Nearby(ctx context.Context, lat float64, lng float64, radiusMeters float64, limit int) ([]Candidate, error)
	IsAvailable(ctx context.Context, driverID string) (bool, error)
	SetLocation(ctx context.Context, driverID string, lat float64, lng float64) error
	SetCooldown(ctx context.Context, driverID string, ttlSeconds int) error
	IsCoolingDown(ctx context.Context, driverID string) (bool, error)
	AcquireRideLock(ctx context.Context, rideID string, ttlSeconds int) (bool, error)
	RefreshRideLock(ctx context.Context, rideID string, ttlSeconds int) error
	ReleaseRideLock(ctx context.Context, rideID string) error
	IncrementOfferCount(ctx context.Context, rideID string, ttlSeconds int) (int, error)
	GetOfferCount(ctx context.Context, rideID string) (int, error)
	HasRideCandidates(ctx context.Context, rideID string) (bool, error)
	SetLastOfferAt(ctx context.Context, driverID string, tsUnix int64) error
	GetLastOfferAt(ctx context.Context, driverIDs []string) (map[string]int64, error)
	StoreRideCandidates(ctx context.Context, rideID string, driverIDs []string, ttlSeconds int) error
	PopRideCandidate(ctx context.Context, rideID string) (string, error)
	SetActiveOffer(ctx context.Context, rideID string, offerID string, driverID string, ttlSeconds int) error
	GetActiveOffer(ctx context.Context, rideID string) (ActiveOffer, bool, error)
	ClearActiveOffer(ctx context.Context, rideID string) error
	ClearRide(ctx context.Context, rideID string) error
}

type ActiveOffer struct {
	OfferID  string
	DriverID string
}
