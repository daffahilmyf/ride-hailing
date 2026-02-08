package grpc

import (
	"context"
	"time"

	userv1 "github.com/daffahilmyf/ride-hailing/proto/user/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type UserClient struct {
	conn   *grpc.ClientConn
	client userv1.UserServiceClient
}

func NewUserClient(addr string, timeout time.Duration) (*UserClient, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	conn, err := grpc.DialContext(ctx, addr, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	if err != nil {
		return nil, err
	}
	return &UserClient{conn: conn, client: userv1.NewUserServiceClient(conn)}, nil
}

func (c *UserClient) Close() error {
	if c == nil || c.conn == nil {
		return nil
	}
	return c.conn.Close()
}

func (c *UserClient) GetUserProfile(ctx context.Context, in *userv1.GetUserProfileRequest, opts ...grpc.CallOption) (*userv1.GetUserProfileResponse, error) {
	return c.client.GetUserProfile(ctx, in, opts...)
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
