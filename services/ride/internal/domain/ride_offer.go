package domain

import (
	"errors"
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

var ErrInvalidOfferTransition = errors.New("invalid offer transition")

func NewRideOffer(rideID, driverID string, ttl time.Duration) RideOffer {
	now := time.Now().UTC()
	id := uuid.NewString()
	return NewRideOfferWith(rideID, driverID, ttl, now, id)
}

func NewRideOfferWith(rideID, driverID string, ttl time.Duration, now time.Time, id string) RideOffer {
	return RideOffer{
		ID:        id,
		RideID:    rideID,
		DriverID:  driverID,
		Status:    OfferPending,
		ExpiresAt: now.UTC().Add(ttl),
		CreatedAt: now.UTC(),
	}
}

func (o RideOffer) Transition(next RideOfferStatus) (RideOffer, error) {
	if o.Status == next {
		return o, nil
	}
	switch o.Status {
	case OfferPending:
		if next == OfferAccepted || next == OfferDeclined || next == OfferExpired {
			o.Status = next
			return o, nil
		}
	case OfferAccepted, OfferDeclined, OfferExpired:
		return o, ErrInvalidOfferTransition
	}
	return o, ErrInvalidOfferTransition
}
