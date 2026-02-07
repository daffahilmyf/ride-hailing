package app

import (
	"context"

	ridev1 "github.com/daffahilmyf/ride-hailing/proto/ride/v1"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type RideServer struct {
	ridev1.UnimplementedRideServiceServer
	logger *zap.Logger
}

func RegisterGRPC(srv *grpc.Server, logger *zap.Logger) {
	ridev1.RegisterRideServiceServer(srv, &RideServer{logger: logger})
}

func (s *RideServer) CreateRide(ctx context.Context, req *ridev1.CreateRideRequest) (*ridev1.CreateRideResponse, error) {
	return nil, status.Error(codes.Unimplemented, "not implemented")
}

func (s *RideServer) StartMatching(ctx context.Context, req *ridev1.StartMatchingRequest) (*ridev1.StartMatchingResponse, error) {
	return nil, status.Error(codes.Unimplemented, "not implemented")
}

func (s *RideServer) AssignDriver(ctx context.Context, req *ridev1.AssignDriverRequest) (*ridev1.AssignDriverResponse, error) {
	return nil, status.Error(codes.Unimplemented, "not implemented")
}

func (s *RideServer) CancelRide(ctx context.Context, req *ridev1.CancelRideRequest) (*ridev1.CancelRideResponse, error) {
	return nil, status.Error(codes.Unimplemented, "not implemented")
}
