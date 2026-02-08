package outbound

import (
	"context"

	userv1 "github.com/daffahilmyf/ride-hailing/proto/user/v1"
	"google.golang.org/grpc"
)

type UserService interface {
	GetUserProfile(ctx context.Context, in *userv1.GetUserProfileRequest, opts ...grpc.CallOption) (*userv1.GetUserProfileResponse, error)
}
