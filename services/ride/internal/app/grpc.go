package app

import (
	"context"
	"errors"

	ridev1 "github.com/daffahilmyf/ride-hailing/proto/ride/v1"
	"github.com/daffahilmyf/ride-hailing/services/ride/internal/app/usecase"
	"github.com/daffahilmyf/ride-hailing/services/ride/internal/domain"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type RideServer struct {
	ridev1.UnimplementedRideServiceServer
	logger  *zap.Logger
	usecase *usecase.RideService
}

func RegisterGRPC(srv *grpc.Server, logger *zap.Logger, uc *usecase.RideService) {
	ridev1.RegisterRideServiceServer(srv, &RideServer{logger: logger, usecase: uc})
}

func (s *RideServer) CreateRide(ctx context.Context, req *ridev1.CreateRideRequest) (*ridev1.CreateRideResponse, error) {
	ride, err := s.usecase.CreateRide(ctx, usecase.CreateRideCmd{
		RiderID:        req.GetRiderId(),
		PickupLat:      req.GetPickupLat(),
		PickupLng:      req.GetPickupLng(),
		DropoffLat:     req.GetDropoffLat(),
		DropoffLng:     req.GetDropoffLng(),
		IdempotencyKey: req.GetIdempotencyKey(),
	})
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to create ride")
	}
	return &ridev1.CreateRideResponse{RideId: ride.ID, Status: string(ride.Status)}, nil
}

func (s *RideServer) StartMatching(ctx context.Context, req *ridev1.StartMatchingRequest) (*ridev1.StartMatchingResponse, error) {
	ride, err := s.usecase.StartMatching(ctx, req.GetRideId())
	if err != nil {
		if errors.Is(err, domain.ErrInvalidTransition) {
			return nil, status.Error(codes.FailedPrecondition, "invalid transition")
		}
		return nil, status.Error(codes.Internal, "failed to start matching")
	}
	return &ridev1.StartMatchingResponse{RideId: ride.ID, Status: string(ride.Status)}, nil
}

func (s *RideServer) AssignDriver(ctx context.Context, req *ridev1.AssignDriverRequest) (*ridev1.AssignDriverResponse, error) {
	ride, err := s.usecase.AssignDriver(ctx, req.GetRideId(), req.GetDriverId())
	if err != nil {
		if errors.Is(err, domain.ErrInvalidTransition) {
			return nil, status.Error(codes.FailedPrecondition, "invalid transition")
		}
		return nil, status.Error(codes.Internal, "failed to assign driver")
	}
	return &ridev1.AssignDriverResponse{RideId: ride.ID, DriverId: req.GetDriverId(), Status: string(ride.Status)}, nil
}

func (s *RideServer) CancelRide(ctx context.Context, req *ridev1.CancelRideRequest) (*ridev1.CancelRideResponse, error) {
	ride, err := s.usecase.CancelRide(ctx, req.GetRideId(), req.GetReason())
	if err != nil {
		if errors.Is(err, domain.ErrInvalidTransition) {
			return nil, status.Error(codes.FailedPrecondition, "invalid transition")
		}
		return nil, status.Error(codes.Internal, "failed to cancel ride")
	}
	return &ridev1.CancelRideResponse{RideId: ride.ID, Status: string(ride.Status)}, nil
}

func (s *RideServer) StartRide(ctx context.Context, req *ridev1.AssignDriverRequest) (*ridev1.AssignDriverResponse, error) {
	ride, err := s.usecase.StartRide(ctx, req.GetRideId())
	if err != nil {
		if errors.Is(err, domain.ErrInvalidTransition) {
			return nil, status.Error(codes.FailedPrecondition, "invalid transition")
		}
		return nil, status.Error(codes.Internal, "failed to start ride")
	}
	return &ridev1.AssignDriverResponse{RideId: ride.ID, DriverId: req.GetDriverId(), Status: string(ride.Status)}, nil
}

func (s *RideServer) CompleteRide(ctx context.Context, req *ridev1.AssignDriverRequest) (*ridev1.AssignDriverResponse, error) {
	ride, err := s.usecase.CompleteRide(ctx, req.GetRideId())
	if err != nil {
		if errors.Is(err, domain.ErrInvalidTransition) {
			return nil, status.Error(codes.FailedPrecondition, "invalid transition")
		}
		return nil, status.Error(codes.Internal, "failed to complete ride")
	}
	return &ridev1.AssignDriverResponse{RideId: ride.ID, DriverId: req.GetDriverId(), Status: string(ride.Status)}, nil
}
