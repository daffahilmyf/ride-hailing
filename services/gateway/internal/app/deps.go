package app

import (
	locationv1 "github.com/daffahilmyf/ride-hailing/proto/location/v1"
	matchingv1 "github.com/daffahilmyf/ride-hailing/proto/matching/v1"
	ridev1 "github.com/daffahilmyf/ride-hailing/proto/ride/v1"
)

type Deps struct {
	RideClient     ridev1.RideServiceClient
	MatchingClient matchingv1.MatchingServiceClient
	LocationClient locationv1.LocationServiceClient
}
