package outbound

import (
	"context"

	ridev1 "github.com/daffahilmyf/ride-hailing/proto/ride/v1"
	"google.golang.org/grpc"
)

type RideService interface {
	CreateRide(ctx context.Context, in *ridev1.CreateRideRequest, opts ...grpc.CallOption) (*ridev1.CreateRideResponse, error)
	CancelRide(ctx context.Context, in *ridev1.CancelRideRequest, opts ...grpc.CallOption) (*ridev1.CancelRideResponse, error)
	CreateOffer(ctx context.Context, in *ridev1.CreateOfferRequest, opts ...grpc.CallOption) (*ridev1.CreateOfferResponse, error)
}
