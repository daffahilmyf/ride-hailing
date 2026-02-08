package handlers

import (
	"context"
	"errors"

	userv1 "github.com/daffahilmyf/ride-hailing/proto/user/v1"
	"github.com/daffahilmyf/ride-hailing/services/user/internal/adapters/db"
	"github.com/daffahilmyf/ride-hailing/services/user/internal/app/metrics"
	"github.com/daffahilmyf/ride-hailing/services/user/internal/app/usecase"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type UserServer struct {
	userv1.UnimplementedUserServiceServer
	logger  *zap.Logger
	usecase *usecase.Service
}

type Dependencies struct {
	Usecase *usecase.Service
	Limiter *RateLimiter
	Metrics *metrics.AuthMetrics
}

func RegisterUserServer(srv *grpc.Server, logger *zap.Logger, deps Dependencies) {
	userv1.RegisterUserServiceServer(srv, &UserServer{logger: logger, usecase: deps.Usecase})
	userv1.RegisterAuthServiceServer(srv, &AuthServer{
		logger:  logger,
		usecase: deps.Usecase,
		limiter: deps.Limiter,
		metrics: deps.Metrics,
	})
}

func (s *UserServer) GetUserProfile(ctx context.Context, req *userv1.GetUserProfileRequest) (*userv1.GetUserProfileResponse, error) {
	user, err := s.usecase.GetUser(ctx, req.GetUserId())
	if err != nil {
		return nil, mapError(err)
	}
	resp := &userv1.GetUserProfileResponse{
		UserId: user.ID,
		Role:   user.Role,
		Name:   user.Name,
	}
	if user.Phone != nil {
		resp.Phone = *user.Phone
	}
	return resp, nil
}

func mapError(err error) error {
	switch {
	case errors.Is(err, db.ErrNotFound):
		return status.Error(codes.NotFound, "user not found")
	case errors.Is(err, usecase.ErrInvalidCredentials):
		return status.Error(codes.InvalidArgument, "invalid request")
	default:
		return status.Error(codes.Internal, "internal error")
	}
}
