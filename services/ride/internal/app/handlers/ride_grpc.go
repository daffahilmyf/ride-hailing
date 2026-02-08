package handlers

import (
	"context"
	"errors"
	"time"

	ridev1 "github.com/daffahilmyf/ride-hailing/proto/ride/v1"
	"github.com/daffahilmyf/ride-hailing/services/ride/internal/app/usecase"
	"github.com/daffahilmyf/ride-hailing/services/ride/internal/domain"
	"github.com/daffahilmyf/ride-hailing/services/ride/internal/ports/outbound"
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

type Dependencies struct {
	Usecase *usecase.RideService
}

func RegisterRideServer(srv *grpc.Server, logger *zap.Logger, deps Dependencies) {
	ridev1.RegisterRideServiceServer(srv, &RideServer{logger: logger, usecase: deps.Usecase})
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
		return nil, mapError(err, "failed to create ride")
	}
	return &ridev1.CreateRideResponse{RideId: ride.ID, Status: string(ride.Status)}, nil
}

func (s *RideServer) StartMatching(ctx context.Context, req *ridev1.StartMatchingRequest) (*ridev1.StartMatchingResponse, error) {
	key := req.GetRequestId()
	ride, err := s.usecase.StartMatching(ctx, req.GetRideId(), key)
	if err != nil {
		return nil, mapError(err, "failed to start matching")
	}
	return &ridev1.StartMatchingResponse{RideId: ride.ID, Status: string(ride.Status)}, nil
}

func (s *RideServer) AssignDriver(ctx context.Context, req *ridev1.AssignDriverRequest) (*ridev1.AssignDriverResponse, error) {
	ride, err := s.usecase.AssignDriver(ctx, req.GetRideId(), req.GetDriverId(), req.GetIdempotencyKey())
	if err != nil {
		return nil, mapError(err, "failed to assign driver")
	}
	return &ridev1.AssignDriverResponse{RideId: ride.ID, DriverId: req.GetDriverId(), Status: string(ride.Status)}, nil
}

func (s *RideServer) CancelRide(ctx context.Context, req *ridev1.CancelRideRequest) (*ridev1.CancelRideResponse, error) {
	key := req.GetRequestId()
	ride, err := s.usecase.CancelRide(ctx, req.GetRideId(), req.GetReason(), key)
	if err != nil {
		return nil, mapError(err, "failed to cancel ride")
	}
	return &ridev1.CancelRideResponse{RideId: ride.ID, Status: string(ride.Status)}, nil
}

func (s *RideServer) StartRide(ctx context.Context, req *ridev1.AssignDriverRequest) (*ridev1.AssignDriverResponse, error) {
	key := req.GetRequestId()
	ride, err := s.usecase.StartRide(ctx, req.GetRideId(), key)
	if err != nil {
		return nil, mapError(err, "failed to start ride")
	}
	return &ridev1.AssignDriverResponse{RideId: ride.ID, DriverId: req.GetDriverId(), Status: string(ride.Status)}, nil
}

func (s *RideServer) CompleteRide(ctx context.Context, req *ridev1.AssignDriverRequest) (*ridev1.AssignDriverResponse, error) {
	key := req.GetRequestId()
	ride, err := s.usecase.CompleteRide(ctx, req.GetRideId(), key)
	if err != nil {
		return nil, mapError(err, "failed to complete ride")
	}
	return &ridev1.AssignDriverResponse{RideId: ride.ID, DriverId: req.GetDriverId(), Status: string(ride.Status)}, nil
}

func (s *RideServer) CreateOffer(ctx context.Context, req *ridev1.CreateOfferRequest) (*ridev1.CreateOfferResponse, error) {
	offer, err := s.usecase.CreateOffer(ctx, usecase.StartMatchingCmd{
		RideID:         req.GetRideId(),
		DriverID:       req.GetDriverId(),
		OfferTTL:       time.Duration(req.GetOfferTtlSeconds()) * time.Second,
		IdempotencyKey: req.GetIdempotencyKey(),
	})
	if err != nil {
		return nil, mapError(err, "failed to create offer")
	}
	return &ridev1.CreateOfferResponse{
		OfferId:   offer.ID,
		RideId:    offer.RideID,
		DriverId:  offer.DriverID,
		Status:    string(offer.Status),
		ExpiresAt: offer.ExpiresAt.Unix(),
	}, nil
}

func (s *RideServer) AcceptOffer(ctx context.Context, req *ridev1.AcceptOfferRequest) (*ridev1.AcceptOfferResponse, error) {
	offer, err := s.usecase.AcceptOffer(ctx, usecase.OfferActionCmd{
		OfferID:        req.GetOfferId(),
		IdempotencyKey: req.GetIdempotencyKey(),
	})
	if err != nil {
		return nil, mapError(err, "failed to accept offer")
	}
	return &ridev1.AcceptOfferResponse{
		OfferId:  offer.ID,
		RideId:   offer.RideID,
		DriverId: offer.DriverID,
		Status:   string(offer.Status),
	}, nil
}

func (s *RideServer) DeclineOffer(ctx context.Context, req *ridev1.DeclineOfferRequest) (*ridev1.DeclineOfferResponse, error) {
	offer, err := s.usecase.DeclineOffer(ctx, usecase.OfferActionCmd{
		OfferID:        req.GetOfferId(),
		IdempotencyKey: req.GetIdempotencyKey(),
	})
	if err != nil {
		return nil, mapError(err, "failed to decline offer")
	}
	return &ridev1.DeclineOfferResponse{
		OfferId:  offer.ID,
		RideId:   offer.RideID,
		DriverId: offer.DriverID,
		Status:   string(offer.Status),
	}, nil
}

func (s *RideServer) ExpireOffer(ctx context.Context, req *ridev1.ExpireOfferRequest) (*ridev1.ExpireOfferResponse, error) {
	offer, err := s.usecase.ExpireOffer(ctx, usecase.OfferActionCmd{
		OfferID:        req.GetOfferId(),
		IdempotencyKey: req.GetIdempotencyKey(),
	})
	if err != nil {
		return nil, mapError(err, "failed to expire offer")
	}
	return &ridev1.ExpireOfferResponse{
		OfferId:  offer.ID,
		RideId:   offer.RideID,
		DriverId: offer.DriverID,
		Status:   string(offer.Status),
	}, nil
}

func mapError(err error, msg string) error {
	switch {
	case errors.Is(err, domain.ErrInvalidTransition):
		return status.Error(codes.FailedPrecondition, "invalid transition")
	case errors.Is(err, outbound.ErrNotFound):
		return status.Error(codes.NotFound, "ride not found")
	case errors.Is(err, outbound.ErrConflict):
		return status.Error(codes.Aborted, "state conflict")
	case errors.Is(err, domain.ErrInvalidOfferTransition):
		return status.Error(codes.FailedPrecondition, "invalid offer transition")
	default:
		return status.Error(codes.Internal, msg)
	}
}
