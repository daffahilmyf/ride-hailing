package outbound

import (
	"context"

	locationv1 "github.com/daffahilmyf/ride-hailing/proto/location/v1"
	"google.golang.org/grpc"
)

type LocationService interface {
	UpdateDriverLocation(ctx context.Context, in *locationv1.UpdateDriverLocationRequest, opts ...grpc.CallOption) (*locationv1.UpdateDriverLocationResponse, error)
	ListNearbyDrivers(ctx context.Context, in *locationv1.ListNearbyDriversRequest, opts ...grpc.CallOption) (*locationv1.ListNearbyDriversResponse, error)
}
