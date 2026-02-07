package domain

import "errors"

type RideStatus string

const (
	StatusRequested      RideStatus = "REQUESTED"
	StatusMatching       RideStatus = "MATCHING"
	StatusOffered        RideStatus = "OFFERED"
	StatusDriverAssigned RideStatus = "DRIVER_ASSIGNED"
	StatusInProgress     RideStatus = "IN_PROGRESS"
	StatusCompleted      RideStatus = "COMPLETED"
	StatusCancelled      RideStatus = "CANCELLED"
)

type Ride struct {
	ID         string
	RiderID    string
	DriverID   *string
	Status     RideStatus
	PickupLat  float64
	PickupLng  float64
	DropoffLat float64
	DropoffLng float64
}

var (
	ErrInvalidTransition = errors.New("invalid state transition")
)

func (r Ride) Transition(next RideStatus) (Ride, error) {
	if r.Status == next {
		return r, nil
	}
	switch r.Status {
	case StatusRequested:
		if next == StatusMatching || next == StatusCancelled {
			r.Status = next
			return r, nil
		}
	case StatusMatching:
		if next == StatusOffered || next == StatusCancelled {
			r.Status = next
			return r, nil
		}
	case StatusOffered:
		if next == StatusDriverAssigned || next == StatusCancelled {
			r.Status = next
			return r, nil
		}
	case StatusDriverAssigned:
		if next == StatusInProgress || next == StatusCancelled {
			r.Status = next
			return r, nil
		}
	case StatusInProgress:
		if next == StatusCompleted {
			r.Status = next
			return r, nil
		}
	}
	return r, ErrInvalidTransition
}
