package domain

import (
	"time"

	"github.com/google/uuid"
)

type RideOfferStatus string

const (
	OfferPending  RideOfferStatus = "PENDING"
	OfferAccepted RideOfferStatus = "ACCEPTED"
	OfferDeclined RideOfferStatus = "DECLINED"
	OfferExpired  RideOfferStatus = "EXPIRED"
)

type RideOffer struct {
	ID        string
	RideID    string
	DriverID  string
	Status    RideOfferStatus
	ExpiresAt time.Time
	CreatedAt time.Time
}

func NewRideOffer(rideID, driverID string, ttl time.Duration) RideOffer {
	now := time.Now().UTC()
	return RideOffer{
		ID:        uuid.NewString(),
		RideID:    rideID,
		DriverID:  driverID,
		Status:    OfferPending,
		ExpiresAt: now.Add(ttl),
		CreatedAt: now,
	}
}
