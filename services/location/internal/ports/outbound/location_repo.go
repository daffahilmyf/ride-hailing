package outbound

import (
	"context"
	"errors"
	"time"
)

var (
	ErrNotFound = errors.New("not found")
)

type Location struct {
	DriverID   string
	Lat        float64
	Lng        float64
	AccuracyM  float64
	RecordedAt time.Time
}

type LocationRepo interface {
	Upsert(ctx context.Context, location Location, ttl time.Duration) error
	Get(ctx context.Context, driverID string) (Location, error)
}
