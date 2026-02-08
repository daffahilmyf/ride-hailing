package app

import "github.com/daffahilmyf/ride-hailing/services/gateway/internal/ports/outbound"

type Deps struct {
	RideClient     outbound.RideService
	MatchingClient outbound.MatchingService
	LocationClient outbound.LocationService
	AuthClient     outbound.AuthService
}
