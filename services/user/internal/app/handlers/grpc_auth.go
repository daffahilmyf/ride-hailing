package handlers

import (
	"context"
	"errors"
	"strings"
	"time"

	userv1 "github.com/daffahilmyf/ride-hailing/proto/user/v1"
	"github.com/daffahilmyf/ride-hailing/services/user/internal/adapters/db"
	"github.com/daffahilmyf/ride-hailing/services/user/internal/app/metrics"
	"github.com/daffahilmyf/ride-hailing/services/user/internal/app/usecase"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AuthServer struct {
	userv1.UnimplementedAuthServiceServer
	logger  *zap.Logger
	usecase *usecase.Service
	limiter *RateLimiter
	metrics *metrics.AuthMetrics
}

func (s *AuthServer) Register(ctx context.Context, req *userv1.RegisterRequest) (*userv1.RegisterResponse, error) {
	start := time.Now()
	if err := s.allow(ctx, "register", req.GetIp()); err != nil {
		s.record("register", "rate_limited", start)
		return nil, err
	}
	user, tokens, code, err := s.usecase.Register(ctx, usecase.RegisterInput{
		Email:     strings.TrimSpace(req.GetEmail()),
		Phone:     strings.TrimSpace(req.GetPhone()),
		Password:  req.GetPassword(),
		Role:      req.GetRole(),
		Name:      req.GetName(),
		DeviceID:  strings.TrimSpace(req.GetDeviceId()),
		UserAgent: req.GetUserAgent(),
		IP:        req.GetIp(),
	})
	if err != nil {
		s.record("register", "error", start)
		return nil, s.mapAuthError(err, "register")
	}
	s.record("register", "ok", start)
	return &userv1.RegisterResponse{
		User:             mapUserProto(user, nil, nil),
		Tokens:           mapTokensProto(tokens),
		VerificationCode: code,
	}, nil
}

func (s *AuthServer) Login(ctx context.Context, req *userv1.LoginRequest) (*userv1.LoginResponse, error) {
	start := time.Now()
	if err := s.allow(ctx, "login", req.GetIp()); err != nil {
		s.record("login", "rate_limited", start)
		return nil, err
	}
	user, tokens, err := s.usecase.Login(ctx, usecase.LoginInput{
		Email:     strings.TrimSpace(req.GetEmail()),
		Phone:     strings.TrimSpace(req.GetPhone()),
		Password:  req.GetPassword(),
		DeviceID:  strings.TrimSpace(req.GetDeviceId()),
		UserAgent: req.GetUserAgent(),
		IP:        req.GetIp(),
	})
	if err != nil {
		s.record("login", "error", start)
		return nil, s.mapAuthError(err, "login")
	}
	s.record("login", "ok", start)
	return &userv1.LoginResponse{
		User:   mapUserProto(user, nil, nil),
		Tokens: mapTokensProto(tokens),
	}, nil
}

func (s *AuthServer) Refresh(ctx context.Context, req *userv1.RefreshRequest) (*userv1.RefreshResponse, error) {
	start := time.Now()
	if err := s.allow(ctx, "refresh", req.GetIp()); err != nil {
		s.record("refresh", "rate_limited", start)
		return nil, err
	}
	user, tokens, err := s.usecase.Refresh(ctx, req.GetRefreshToken(), strings.TrimSpace(req.GetDeviceId()))
	if err != nil {
		s.record("refresh", "error", start)
		return nil, s.mapAuthError(err, "refresh")
	}
	s.record("refresh", "ok", start)
	return &userv1.RefreshResponse{
		User:   mapUserProto(user, nil, nil),
		Tokens: mapTokensProto(tokens),
	}, nil
}

func (s *AuthServer) Logout(ctx context.Context, req *userv1.LogoutRequest) (*userv1.LogoutResponse, error) {
	start := time.Now()
	if err := s.allow(ctx, "logout", req.GetDeviceId()); err != nil {
		s.record("logout", "rate_limited", start)
		return nil, err
	}
	if err := s.usecase.Logout(ctx, req.GetRefreshToken(), strings.TrimSpace(req.GetDeviceId())); err != nil {
		s.record("logout", "error", start)
		return nil, s.mapAuthError(err, "logout")
	}
	s.record("logout", "ok", start)
	return &userv1.LogoutResponse{Status: "ok"}, nil
}

func (s *AuthServer) Verify(ctx context.Context, req *userv1.VerifyRequest) (*userv1.VerifyResponse, error) {
	start := time.Now()
	if err := s.allow(ctx, "verify", req.GetTarget()); err != nil {
		s.record("verify", "rate_limited", start)
		return nil, err
	}
	if err := s.usecase.Verify(ctx, req.GetChannel(), req.GetTarget(), req.GetCode()); err != nil {
		s.record("verify", "error", start)
		return nil, s.mapAuthError(err, "verify")
	}
	s.record("verify", "ok", start)
	return &userv1.VerifyResponse{Status: "ok"}, nil
}

func (s *AuthServer) LogoutAll(ctx context.Context, req *userv1.LogoutAllRequest) (*userv1.LogoutAllResponse, error) {
	start := time.Now()
	if err := s.usecase.LogoutAll(ctx, req.GetUserId()); err != nil {
		s.record("logout_all", "error", start)
		return nil, status.Error(codes.Internal, "internal")
	}
	s.record("logout_all", "ok", start)
	return &userv1.LogoutAllResponse{Status: "ok"}, nil
}

func (s *AuthServer) LogoutDevice(ctx context.Context, req *userv1.LogoutDeviceRequest) (*userv1.LogoutDeviceResponse, error) {
	start := time.Now()
	if strings.TrimSpace(req.GetDeviceId()) == "" {
		s.record("logout_device", "bad_request", start)
		return nil, status.Error(codes.InvalidArgument, "invalid_request")
	}
	if err := s.usecase.Repo.RevokeDeviceSessions(ctx, req.GetUserId(), strings.TrimSpace(req.GetDeviceId())); err != nil {
		s.record("logout_device", "error", start)
		return nil, status.Error(codes.Internal, "internal")
	}
	s.record("logout_device", "ok", start)
	return &userv1.LogoutDeviceResponse{Status: "ok"}, nil
}

func (s *AuthServer) ListSessions(ctx context.Context, req *userv1.ListSessionsRequest) (*userv1.ListSessionsResponse, error) {
	start := time.Now()
	tokens, err := s.usecase.Repo.ListActiveDeviceSessions(ctx, req.GetUserId())
	if err != nil {
		s.record("sessions", "error", start)
		return nil, status.Error(codes.Internal, "internal")
	}
	sessions := make([]*userv1.Session, 0, len(tokens))
	for _, t := range tokens {
		sessions = append(sessions, &userv1.Session{
			DeviceId:  t.DeviceID,
			UserAgent: t.UserAgent,
			Ip:        t.IP,
			CreatedAt: t.CreatedAt.UTC().Format(time.RFC3339),
			ExpiresAt: t.ExpiresAt.UTC().Format(time.RFC3339),
		})
	}
	s.record("sessions", "ok", start)
	return &userv1.ListSessionsResponse{Sessions: sessions}, nil
}

func (s *AuthServer) GetMe(ctx context.Context, req *userv1.GetMeRequest) (*userv1.GetMeResponse, error) {
	start := time.Now()
	user, err := s.usecase.GetUser(ctx, req.GetUserId())
	if err != nil {
		s.record("me", "error", start)
		if errors.Is(err, db.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "not_found")
		}
		return nil, status.Error(codes.Internal, "internal")
	}
	var rider *db.RiderProfile
	var driver *db.DriverProfile
	if user.Role == "rider" {
		if prof, err := s.usecase.Repo.GetRiderProfile(ctx, user.ID); err == nil {
			rider = &prof
		}
	} else if user.Role == "driver" {
		if prof, err := s.usecase.Repo.GetDriverProfile(ctx, user.ID); err == nil {
			driver = &prof
		}
	}
	s.record("me", "ok", start)
	return &userv1.GetMeResponse{User: mapUserProto(user, rider, driver)}, nil
}

func (s *AuthServer) allow(ctx context.Context, endpoint string, key string) error {
	if s.limiter == nil || s.limiter.Redis == nil {
		return nil
	}
	if key == "" {
		key = "unknown"
	}
	allowed, err := s.limiter.Redis.Allow(ctx, s.limiter.Prefix+key+":"+endpoint, s.limiter.Limit, s.limiter.Window)
	if err == nil && allowed {
		return nil
	}
	if s.limiter.OnLimit != nil {
		s.limiter.OnLimit(endpoint)
	}
	return status.Error(codes.ResourceExhausted, "rate_limited")
}

func (s *AuthServer) record(endpoint string, status string, start time.Time) {
	if s.metrics != nil {
		s.metrics.Record(endpoint, status, time.Since(start))
	}
}

func (s *AuthServer) mapAuthError(err error, endpoint string) error {
	switch {
	case errors.Is(err, usecase.ErrWeakPassword):
		return status.Error(codes.InvalidArgument, "weak_password")
	case errors.Is(err, usecase.ErrInvalidRole):
		return status.Error(codes.InvalidArgument, "invalid_request")
	case errors.Is(err, usecase.ErrAlreadyExists):
		return status.Error(codes.AlreadyExists, "already_exists")
	case errors.Is(err, usecase.ErrDeviceRequired):
		return status.Error(codes.InvalidArgument, "device_required")
	case errors.Is(err, usecase.ErrAccountLocked):
		return status.Error(codes.ResourceExhausted, "account_locked")
	case errors.Is(err, usecase.ErrInvalidCredentials):
		return status.Error(codes.Unauthenticated, authInvalidMessage(endpoint))
	default:
		return status.Error(codes.Internal, "internal")
	}
}

func authInvalidMessage(endpoint string) string {
	switch endpoint {
	case "refresh", "logout":
		return "invalid_refresh"
	case "verify":
		return "invalid_code"
	default:
		return "invalid_credentials"
	}
}

func mapTokensProto(tokens usecase.Tokens) *userv1.Token {
	return &userv1.Token{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		TokenType:    "bearer",
		ExpiresIn:    tokens.ExpiresIn,
	}
}

func mapUserProto(user db.User, rider *db.RiderProfile, driver *db.DriverProfile) *userv1.User {
	resp := &userv1.User{
		Id:   user.ID,
		Role: user.Role,
		Name: user.Name,
	}
	if user.Email != nil {
		resp.Email = *user.Email
	}
	if user.Phone != nil {
		resp.Phone = *user.Phone
	}
	if user.EmailVerifiedAt != nil {
		resp.EmailVerifiedAt = user.EmailVerifiedAt.UTC().Format(time.RFC3339)
	}
	if user.PhoneVerifiedAt != nil {
		resp.PhoneVerifiedAt = user.PhoneVerifiedAt.UTC().Format(time.RFC3339)
	}
	if rider != nil {
		resp.RiderProfile = &userv1.RiderProfile{
			Rating:            rider.Rating,
			PreferredLanguage: rider.PreferredLanguage,
		}
	}
	if driver != nil {
		resp.DriverProfile = &userv1.DriverProfile{
			VehicleMake:   driver.VehicleMake,
			VehiclePlate:  driver.VehiclePlate,
			LicenseNumber: driver.LicenseNumber,
			Verified:      driver.Verified,
			Rating:        driver.Rating,
		}
	}
	return resp
}
