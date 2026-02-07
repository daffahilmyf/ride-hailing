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
	UpdateStatus(ctx context.Context, id string, status string) error
}
