package grpc

import (
	"context"
	"time"

	locationv1 "github.com/daffahilmyf/ride-hailing/proto/location/v1"
	matchingv1 "github.com/daffahilmyf/ride-hailing/proto/matching/v1"
	ridev1 "github.com/daffahilmyf/ride-hailing/proto/ride/v1"
	userv1 "github.com/daffahilmyf/ride-hailing/proto/user/v1"
	"github.com/daffahilmyf/ride-hailing/services/gateway/internal/infra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type Clients struct {
	RideConn     *grpc.ClientConn
	MatchingConn *grpc.ClientConn
	LocationConn *grpc.ClientConn
	UserConn     *grpc.ClientConn

	RideClient     ridev1.RideServiceClient
	MatchingClient matchingv1.MatchingServiceClient
	LocationClient locationv1.LocationServiceClient
	AuthClient     userv1.AuthServiceClient
}

func NewClients(ctx context.Context, cfg infra.GRPCConfig) (*Clients, error) {
	dialOpts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithChainUnaryInterceptor(
			timeoutUnaryInterceptor(time.Duration(cfg.TimeoutSeconds)*time.Second),
			retryUnaryInterceptor(cfg.RetryMax, time.Duration(cfg.RetryBackoffMs)*time.Millisecond),
		),
	}

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
	userConn, err := dialWithTimeout(ctx, cfg.UserAddr, dialOpts...)
	if err != nil {
		_ = rideConn.Close()
		_ = matchingConn.Close()
		_ = locationConn.Close()
		return nil, err
	}

	return &Clients{
		RideConn:       rideConn,
		MatchingConn:   matchingConn,
		LocationConn:   locationConn,
		UserConn:       userConn,
		RideClient:     ridev1.NewRideServiceClient(rideConn),
		MatchingClient: matchingv1.NewMatchingServiceClient(matchingConn),
		LocationClient: locationv1.NewLocationServiceClient(locationConn),
		AuthClient:     userv1.NewAuthServiceClient(userConn),
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
	if c.UserConn != nil {
		_ = c.UserConn.Close()
	}
	return nil
}

func dialWithTimeout(ctx context.Context, addr string, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	return grpc.DialContext(ctx, addr, opts...)
}

func timeoutUnaryInterceptor(timeout time.Duration) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		if timeout <= 0 {
			return invoker(ctx, method, req, reply, cc, opts...)
		}
		if _, ok := ctx.Deadline(); ok {
			return invoker(ctx, method, req, reply, cc, opts...)
		}
		tctx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()
		return invoker(tctx, method, req, reply, cc, opts...)
	}
}

func retryUnaryInterceptor(maxRetries int, backoff time.Duration) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		var err error
		attempts := maxRetries + 1
		if attempts < 1 {
			attempts = 1
		}
		for i := 0; i < attempts; i++ {
			err = invoker(ctx, method, req, reply, cc, opts...)
			if err == nil || !isRetryable(err) {
				return err
			}
			if backoff > 0 {
				time.Sleep(backoff)
			}
		}
		return err
	}
}

func isRetryable(err error) bool {
	st, ok := status.FromError(err)
	if !ok {
		return false
	}
	switch st.Code() {
	case codes.Unavailable, codes.DeadlineExceeded:
		return true
	default:
		return false
	}
}

func WithInternalToken(ctx context.Context, token string) context.Context {
	if token == "" {
		return ctx
	}
	md := metadata.Pairs("x-internal-token", token)
	return metadata.NewOutgoingContext(ctx, md)
}
