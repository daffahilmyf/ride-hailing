package outbound

import (
	"context"

	userv1 "github.com/daffahilmyf/ride-hailing/proto/user/v1"
	"google.golang.org/grpc"
)

type AuthService interface {
	Register(ctx context.Context, in *userv1.RegisterRequest, opts ...grpc.CallOption) (*userv1.RegisterResponse, error)
	Login(ctx context.Context, in *userv1.LoginRequest, opts ...grpc.CallOption) (*userv1.LoginResponse, error)
	Refresh(ctx context.Context, in *userv1.RefreshRequest, opts ...grpc.CallOption) (*userv1.RefreshResponse, error)
	Logout(ctx context.Context, in *userv1.LogoutRequest, opts ...grpc.CallOption) (*userv1.LogoutResponse, error)
	Verify(ctx context.Context, in *userv1.VerifyRequest, opts ...grpc.CallOption) (*userv1.VerifyResponse, error)
	LogoutAll(ctx context.Context, in *userv1.LogoutAllRequest, opts ...grpc.CallOption) (*userv1.LogoutAllResponse, error)
	LogoutDevice(ctx context.Context, in *userv1.LogoutDeviceRequest, opts ...grpc.CallOption) (*userv1.LogoutDeviceResponse, error)
	ListSessions(ctx context.Context, in *userv1.ListSessionsRequest, opts ...grpc.CallOption) (*userv1.ListSessionsResponse, error)
	GetMe(ctx context.Context, in *userv1.GetMeRequest, opts ...grpc.CallOption) (*userv1.GetMeResponse, error)
}
