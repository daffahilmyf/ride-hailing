package grpc

import (
	"github.com/daffahilmyf/ride-hailing/services/location/internal/app/handlers"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type Server struct {
	grpc *grpc.Server
}

func NewServer(logger *zap.Logger, deps handlers.Dependencies, metrics *Metrics, auth AuthConfig) *Server {
	server := grpc.NewServer(
		grpc.ChainUnaryInterceptor(UnaryInterceptors(logger, metrics, auth)...),
	)
	handlers.RegisterLocationServer(server, logger, deps)
	return &Server{grpc: server}
}

func (s *Server) GRPC() *grpc.Server {
	return s.grpc
}
