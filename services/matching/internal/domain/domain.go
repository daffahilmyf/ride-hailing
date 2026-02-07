package domain

import "errors"

var (
	ErrInvalidStatus = errors.New("invalid status")
)

type DriverStatus string

const (
	StatusOffline DriverStatus = "OFFLINE"
	StatusOnline  DriverStatus = "ONLINE_AVAILABLE"
	StatusOnTrip  DriverStatus = "ON_TRIP"
	StatusOffered DriverStatus = "OFFERED"
)

func ParseStatus(value string) (DriverStatus, error) {
	switch DriverStatus(value) {
	case StatusOffline, StatusOnline, StatusOnTrip, StatusOffered:
		return DriverStatus(value), nil
	default:
		return "", ErrInvalidStatus
	}
}

type Candidate struct {
	DriverID  string
	DistanceM float64
}
