package handlers

import (
	"context"
	"errors"

	locationv1 "github.com/daffahilmyf/ride-hailing/proto/location/v1"
	"github.com/daffahilmyf/ride-hailing/services/location/internal/app/usecase"
	"github.com/daffahilmyf/ride-hailing/services/location/internal/domain"
	"github.com/daffahilmyf/ride-hailing/services/location/internal/ports/outbound"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type LocationServer struct {
	locationv1.UnimplementedLocationServiceServer
	logger  *zap.Logger
	usecase *usecase.LocationService
}

type Dependencies struct {
	Usecase *usecase.LocationService
}

func RegisterLocationServer(srv *grpc.Server, logger *zap.Logger, deps Dependencies) {
	locationv1.RegisterLocationServiceServer(srv, &LocationServer{logger: logger, usecase: deps.Usecase})
}

func (s *LocationServer) GetDriverLocation(ctx context.Context, req *locationv1.GetDriverLocationRequest) (*locationv1.GetDriverLocationResponse, error) {
	location, err := s.usecase.GetDriverLocation(ctx, req.GetDriverId())
	if err != nil {
		return nil, mapError(err, "failed to get driver location")
	}
	return &locationv1.GetDriverLocationResponse{
		DriverId:       location.DriverID,
		Lat:            location.Lat,
		Lng:            location.Lng,
		RecordedAtUnix: location.RecordedAt.Unix(),
	}, nil
}

func (s *LocationServer) UpdateDriverLocation(ctx context.Context, req *locationv1.UpdateDriverLocationRequest) (*locationv1.UpdateDriverLocationResponse, error) {
	_, err := s.usecase.UpdateDriverLocation(ctx, req.GetDriverId(), req.GetLat(), req.GetLng(), req.GetAccuracyM())
	if err != nil {
		return nil, mapError(err, "failed to update driver location")
	}
	return &locationv1.UpdateDriverLocationResponse{Status: "OK"}, nil
}

func mapError(err error, msg string) error {
	switch {
	case errors.Is(err, domain.ErrInvalidLocation):
		return status.Error(codes.InvalidArgument, "invalid location")
	case errors.Is(err, outbound.ErrNotFound):
		return status.Error(codes.NotFound, "location not found")
	default:
		return status.Error(codes.Internal, msg)
	}
}
