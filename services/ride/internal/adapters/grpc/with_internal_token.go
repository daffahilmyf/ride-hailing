package grpc

import (
	"context"

	"google.golang.org/grpc/metadata"
)

func WithInternalToken(ctx context.Context, token string) context.Context {
	if token == "" {
		return ctx
	}
	return metadata.AppendToOutgoingContext(ctx, "x-internal-token", token)
}
