package outbound

import (
	"context"
	"time"
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
	UpdateStatus(ctx context.Context, id string, status string, updatedAt time.Time) error
	AssignDriver(ctx context.Context, id string, driverID string, status string, updatedAt time.Time) error
}
