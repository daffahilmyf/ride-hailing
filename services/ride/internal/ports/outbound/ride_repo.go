package outbound

import (
	"context"
	"errors"
	"time"
)

var (
	ErrNotFound = errors.New("not found")
	ErrConflict = errors.New("conflict")
)

type Ride struct {
	ID         string
	RiderID    string
	DriverID   *string
	Status     string
	PickupLat  float64
	PickupLng  float64
	DropoffLat float64
	DropoffLng float64
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type RideRepo interface {
	Create(ctx context.Context, ride Ride) error
	Get(ctx context.Context, id string) (Ride, error)
	UpdateStatusIfCurrent(ctx context.Context, id string, currentStatus string, nextStatus string, updatedAt time.Time) error
	AssignDriverIfCurrent(ctx context.Context, id string, driverID string, currentStatus string, nextStatus string, updatedAt time.Time) error
}
