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

func (s *LocationServer) ListNearbyDrivers(ctx context.Context, req *locationv1.ListNearbyDriversRequest) (*locationv1.ListNearbyDriversResponse, error) {
	radius := req.GetRadiusM()
	if radius <= 0 {
		return nil, status.Error(codes.InvalidArgument, "radius must be > 0")
	}
	limit := int(req.GetLimit())
	if limit <= 0 {
		limit = 10
	}
	drivers, err := s.usecase.ListNearbyDrivers(ctx, req.GetLat(), req.GetLng(), radius, limit)
	if err != nil {
		return nil, mapError(err, "failed to list nearby drivers")
	}
	resp := &locationv1.ListNearbyDriversResponse{
		Drivers: make([]*locationv1.NearbyDriver, 0, len(drivers)),
	}
	for _, driver := range drivers {
		resp.Drivers = append(resp.Drivers, &locationv1.NearbyDriver{
			DriverId:  driver.DriverID,
			Lat:       driver.Lat,
			Lng:       driver.Lng,
			DistanceM: driver.DistanceM,
		})
	}
	return resp, nil
}

func mapError(err error, msg string) error {
	switch {
	case errors.Is(err, domain.ErrInvalidLocation):
		return status.Error(codes.InvalidArgument, "invalid location")
	case errors.Is(err, domain.ErrRateLimited):
		return status.Error(codes.ResourceExhausted, "rate limited")
	case errors.Is(err, outbound.ErrNotFound):
		return status.Error(codes.NotFound, "location not found")
	default:
		return status.Error(codes.Internal, msg)
	}
}
