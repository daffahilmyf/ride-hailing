package grpc

import (
	"context"
	"time"

	"github.com/daffahilmyf/ride-hailing/services/gateway/internal/infra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Clients struct {
	RideConn     *grpc.ClientConn
	MatchingConn *grpc.ClientConn
	LocationConn *grpc.ClientConn
}

func NewClients(ctx context.Context, cfg infra.GRPCConfig) (*Clients, error) {
	dialOpts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}

	rideConn, err := dialWithTimeout(ctx, cfg.RideAddr, dialOpts...)
	if err != nil {
		return nil, err
	}
	matchingConn, err := dialWithTimeout(ctx, cfg.MatchingAddr, dialOpts...)
	if err != nil {
		_ = rideConn.Close()
		return nil, err
	}
	locationConn, err := dialWithTimeout(ctx, cfg.LocationAddr, dialOpts...)
	if err != nil {
		_ = rideConn.Close()
		_ = matchingConn.Close()
		return nil, err
	}

	return &Clients{
		RideConn:     rideConn,
		MatchingConn: matchingConn,
		LocationConn: locationConn,
	}, nil
}

func (c *Clients) Close() error {
	if c == nil {
		return nil
	}
	if c.RideConn != nil {
		_ = c.RideConn.Close()
	}
	if c.MatchingConn != nil {
		_ = c.MatchingConn.Close()
	}
	if c.LocationConn != nil {
		_ = c.LocationConn.Close()
	}
	return nil
}

func dialWithTimeout(ctx context.Context, addr string, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	return grpc.DialContext(ctx, addr, opts...)
}
