package grpcadapter

import (
	"net"

	"github.com/daffahilmyf/ride-hailing/services/user/internal/app/handlers"
	"github.com/daffahilmyf/ride-hailing/services/user/internal/app/metrics"
	"github.com/daffahilmyf/ride-hailing/services/user/internal/app/usecase"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	health "google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

type Server struct {
	grpc *grpc.Server
}

type Dependencies struct {
	Usecase *usecase.Service
	Limiter *handlers.RateLimiter
	Metrics *metrics.AuthMetrics
}

func NewServer(logger *zap.Logger, deps Dependencies, auth AuthConfig) *Server {
	srv := grpc.NewServer(
		grpc.ChainUnaryInterceptor(UnaryInterceptors(logger, auth)...),
	)
	handlers.RegisterUserServer(srv, logger, handlers.Dependencies{
		Usecase: deps.Usecase,
		Limiter: deps.Limiter,
		Metrics: deps.Metrics,
	})
	healthSrv := health.NewServer()
	healthpb.RegisterHealthServer(srv, healthSrv)
	reflection.Register(srv)
	return &Server{grpc: srv}
}

func (s *Server) Serve(addr string) error {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	return s.grpc.Serve(lis)
}

func (s *Server) GRPC() *grpc.Server {
	return s.grpc
}
