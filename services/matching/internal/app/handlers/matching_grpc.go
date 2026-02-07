package handlers

import (
	"context"
	"errors"

	matchingv1 "github.com/daffahilmyf/ride-hailing/proto/matching/v1"
	"github.com/daffahilmyf/ride-hailing/services/matching/internal/app/usecase"
	"github.com/daffahilmyf/ride-hailing/services/matching/internal/domain"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type MatchingServer struct {
	matchingv1.UnimplementedMatchingServiceServer
	logger  *zap.Logger
	usecase *usecase.MatchingService
}

type Dependencies struct {
	Usecase *usecase.MatchingService
}

func RegisterMatchingServer(srv *grpc.Server, logger *zap.Logger, deps Dependencies) {
	matchingv1.RegisterMatchingServiceServer(srv, &MatchingServer{logger: logger, usecase: deps.Usecase})
}

func (s *MatchingServer) UpdateDriverStatus(ctx context.Context, req *matchingv1.UpdateDriverStatusRequest) (*matchingv1.UpdateDriverStatusResponse, error) {
	if err := s.usecase.UpdateDriverStatus(ctx, req.GetDriverId(), req.GetStatus()); err != nil {
		return nil, mapError(err, "failed to update status")
	}
	return &matchingv1.UpdateDriverStatusResponse{Status: "OK"}, nil
}

func (s *MatchingServer) FindCandidates(ctx context.Context, req *matchingv1.FindCandidatesRequest) (*matchingv1.FindCandidatesResponse, error) {
	candidates, err := s.usecase.FindCandidates(ctx, req.GetPickupLat(), req.GetPickupLng(), 0, int(req.GetLimit()))
	if err != nil {
		return nil, mapError(err, "failed to find candidates")
	}
	resp := &matchingv1.FindCandidatesResponse{
		Candidates: make([]*matchingv1.Candidate, 0, len(candidates)),
	}
	for _, candidate := range candidates {
		resp.Candidates = append(resp.Candidates, &matchingv1.Candidate{
			DriverId:  candidate.DriverID,
			DistanceM: candidate.DistanceM,
		})
	}
	return resp, nil
}

func (s *MatchingServer) NotifyOfferSent(ctx context.Context, req *matchingv1.NotifyOfferSentRequest) (*matchingv1.NotifyOfferSentResponse, error) {
	if err := s.usecase.NotifyOfferSent(ctx, req.GetDriverId(), req.GetOfferId()); err != nil {
		return nil, mapError(err, "failed to notify offer")
	}
	return &matchingv1.NotifyOfferSentResponse{Status: "OK"}, nil
}

func mapError(err error, msg string) error {
	switch {
	case errors.Is(err, domain.ErrInvalidStatus):
		return status.Error(codes.InvalidArgument, "invalid status")
	default:
		return status.Error(codes.Internal, msg)
	}
}
