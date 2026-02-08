package domain

import (
	"errors"
	"time"
)

var (
	ErrInvalidLocation = errors.New("invalid location")
)

type DriverLocation struct {
	DriverID   string
	Lat        float64
	Lng        float64
	AccuracyM  float64
	RecordedAt time.Time
}

func NewDriverLocation(driverID string, lat float64, lng float64, accuracy float64, recordedAt time.Time) (DriverLocation, error) {
	if driverID == "" {
		return DriverLocation{}, ErrInvalidLocation
	}
	if lat < -90 || lat > 90 || lng < -180 || lng > 180 {
		return DriverLocation{}, ErrInvalidLocation
	}
	if accuracy < 0 {
		return DriverLocation{}, ErrInvalidLocation
	}
	if recordedAt.IsZero() {
		return DriverLocation{}, ErrInvalidLocation
	}
	return DriverLocation{
		DriverID:   driverID,
		Lat:        lat,
		Lng:        lng,
		AccuracyM:  accuracy,
		RecordedAt: recordedAt.UTC(),
	}, nil
}
