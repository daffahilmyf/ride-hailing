package outbound

import "context"

type RideOffer struct {
	ID        string
	RideID    string
	DriverID  string
	Status    string
	ExpiresAt int64
	CreatedAt int64
}

type RideOfferRepo interface {
	Create(ctx context.Context, offer RideOffer) error
	Get(ctx context.Context, id string) (RideOffer, error)
	UpdateStatusIfCurrent(ctx context.Context, id string, currentStatus string, nextStatus string) error
	ListExpired(ctx context.Context, cutoff int64, limit int) ([]RideOffer, error)
}
