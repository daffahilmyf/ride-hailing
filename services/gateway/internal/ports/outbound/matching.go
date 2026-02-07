package outbound

import (
	"context"

	matchingv1 "github.com/daffahilmyf/ride-hailing/proto/matching/v1"
	"google.golang.org/grpc"
)

type MatchingService interface {
	UpdateDriverStatus(ctx context.Context, in *matchingv1.UpdateDriverStatusRequest, opts ...grpc.CallOption) (*matchingv1.UpdateDriverStatusResponse, error)
}
