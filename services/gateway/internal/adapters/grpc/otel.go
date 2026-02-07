package grpc

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"google.golang.org/grpc/metadata"
)

type mdCarrier metadata.MD

func (c mdCarrier) Get(key string) string {
	vals := metadata.MD(c).Get(key)
	if len(vals) == 0 {
		return ""
	}
	return vals[0]
}

func (c mdCarrier) Set(key string, value string) {
	metadata.MD(c).Set(key, value)
}

func (c mdCarrier) Keys() []string {
	keys := make([]string, 0, len(c))
	for k := range c {
		keys = append(keys, k)
	}
	return keys
}

func WithTraceContext(ctx context.Context) context.Context {
	md := metadata.New(nil)
	otel.GetTextMapPropagator().Inject(ctx, propagation.TextMapCarrier(mdCarrier(md)))
	return metadata.NewOutgoingContext(ctx, md)
}
