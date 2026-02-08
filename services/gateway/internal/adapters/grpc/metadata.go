package grpc

import (
	"context"

	"google.golang.org/grpc/metadata"
)

func WithRequestMetadata(ctx context.Context, traceID, requestID string) context.Context {
	return metadata.AppendToOutgoingContext(ctx,
		"x-trace-id", traceID,
		"x-request-id", requestID,
	)
}
