package grpc

import (
	"context"
	"errors"
	"math"
	"math/rand"
	"sync"
	"time"

	userv1 "github.com/daffahilmyf/ride-hailing/proto/user/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

type UserClient struct {
	conn         *grpc.ClientConn
	client       userv1.UserServiceClient
	breaker      *CircuitBreaker
	timeout      time.Duration
	retryMax     int
	retryBackoff time.Duration
}

func NewUserClient(addr string, timeout time.Duration, breaker *CircuitBreaker, requestTimeout time.Duration, retryMax int, retryBackoff time.Duration) (*UserClient, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	conn, err := grpc.DialContext(ctx, addr, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	if err != nil {
		return nil, err
	}
	return &UserClient{
		conn:         conn,
		client:       userv1.NewUserServiceClient(conn),
		breaker:      breaker,
		timeout:      requestTimeout,
		retryMax:     retryMax,
		retryBackoff: retryBackoff,
	}, nil
}

func (c *UserClient) Close() error {
	if c == nil || c.conn == nil {
		return nil
	}
	return c.conn.Close()
}

func (c *UserClient) GetUserProfile(ctx context.Context, in *userv1.GetUserProfileRequest, opts ...grpc.CallOption) (*userv1.GetUserProfileResponse, error) {
	call := func(callCtx context.Context) (*userv1.GetUserProfileResponse, error) {
		if c.breaker == nil {
			return c.client.GetUserProfile(callCtx, in, opts...)
		}
		res, err := c.breaker.Execute(func() (any, error) {
			return c.client.GetUserProfile(callCtx, in, opts...)
		})
		if err == nil {
			return res.(*userv1.GetUserProfileResponse), nil
		}
		if errors.Is(err, ErrCircuitOpen) {
			return nil, status.Error(codes.Unavailable, "user service circuit open")
		}
		return nil, err
	}

	return c.callWithRetry(ctx, call)
}

func (c *UserClient) callWithRetry(ctx context.Context, call func(context.Context) (*userv1.GetUserProfileResponse, error)) (*userv1.GetUserProfileResponse, error) {
	attempts := 1 + c.retryMax
	if attempts < 1 {
		attempts = 1
	}

	var lastErr error
	for attempt := 0; attempt < attempts; attempt++ {
		callCtx, cancel := c.withTimeout(ctx)
		resp, err := call(callCtx)
		if cancel != nil {
			cancel()
		}
		if err == nil {
			return resp, nil
		}
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return nil, err
		}
		if !isRetryable(err) || attempt == attempts-1 {
			return nil, err
		}
		lastErr = err
		time.Sleep(backoffForAttempt(c.retryBackoff, attempt))
	}
	return nil, lastErr
}

func (c *UserClient) withTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	if c.timeout <= 0 {
		return ctx, nil
	}
	if _, hasDeadline := ctx.Deadline(); hasDeadline {
		return ctx, nil
	}
	return context.WithTimeout(ctx, c.timeout)
}

func isRetryable(err error) bool {
	code := status.Code(err)
	switch code {
	case codes.Unavailable, codes.DeadlineExceeded, codes.ResourceExhausted:
		return true
	default:
		return false
	}
}

func backoffForAttempt(base time.Duration, attempt int) time.Duration {
	if base <= 0 {
		return 0
	}
	pow := math.Pow(2, float64(attempt))
	backoff := time.Duration(pow) * base
	max := 2 * time.Second
	if backoff > max {
		return max
	}
	return addJitter(backoff)
}

var jitterRand = rand.New(rand.NewSource(time.Now().UnixNano()))
var jitterMu sync.Mutex

func addJitter(d time.Duration) time.Duration {
	if d <= 0 {
		return d
	}
	jitterMu.Lock()
	defer jitterMu.Unlock()
	maxJitter := d / 2
	if maxJitter <= 0 {
		return d
	}
	jitter := time.Duration(jitterRand.Int63n(int64(maxJitter) + 1))
	return d + jitter
}

type UserClientWithToken struct {
	inner *UserClient
	token string
}

func NewUserClientWithToken(inner *UserClient, token string) *UserClientWithToken {
	return &UserClientWithToken{inner: inner, token: token}
}

func (c *UserClientWithToken) GetUserProfile(ctx context.Context, in *userv1.GetUserProfileRequest, opts ...grpc.CallOption) (*userv1.GetUserProfileResponse, error) {
	ctx = WithInternalToken(ctx, c.token)
	return c.inner.GetUserProfile(ctx, in, opts...)
}
