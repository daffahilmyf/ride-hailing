package grpc

import (
	"github.com/daffahilmyf/ride-hailing/services/ride/internal/app/handlers"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type Server struct {
	grpc *grpc.Server
}

func NewServer(logger *zap.Logger, deps handlers.Dependencies, metrics *Metrics) *Server {
	srv := grpc.NewServer(
		grpc.ChainUnaryInterceptor(UnaryInterceptors(logger, metrics)...),
	)
	handlers.RegisterRideServer(srv, logger, deps)
	return &Server{grpc: srv}
}

func (s *Server) GRPC() *grpc.Server {
	return s.grpc
}
