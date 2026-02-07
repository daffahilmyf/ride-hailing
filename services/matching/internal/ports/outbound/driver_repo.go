package outbound

import "context"

type Candidate struct {
	DriverID  string
	DistanceM float64
}

type DriverRepo interface {
	UpdateStatus(ctx context.Context, driverID string, status string) error
	MarkOfferSent(ctx context.Context, driverID string, offerID string, ttlSeconds int) error
	Nearby(ctx context.Context, lat float64, lng float64, radiusMeters float64, limit int) ([]Candidate, error)
	IsAvailable(ctx context.Context, driverID string) (bool, error)
	SetLocation(ctx context.Context, driverID string, lat float64, lng float64) error
}
