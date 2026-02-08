package grpcadapter

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type ctxKey string

const (
	ctxKeyTraceID   ctxKey = "trace_id"
	ctxKeyRequestID ctxKey = "request_id"
)

type AuthConfig struct {
	Enabled bool
	Token   string
}

func UnaryInterceptors(logger *zap.Logger, auth AuthConfig) []grpc.UnaryServerInterceptor {
	return []grpc.UnaryServerInterceptor{
		recoveryInterceptor(logger),
		requestIDInterceptor(),
		authInterceptor(auth),
		loggingInterceptor(logger),
	}
}

func TraceIDFromContext(ctx context.Context) string {
	if v, ok := ctx.Value(ctxKeyTraceID).(string); ok {
		return v
	}
	return ""
}

func RequestIDFromContext(ctx context.Context) string {
	if v, ok := ctx.Value(ctxKeyRequestID).(string); ok {
		return v
	}
	return ""
}

func requestIDInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		md, _ := metadata.FromIncomingContext(ctx)
		traceID := firstHeader(md, "x-trace-id")
		if traceID == "" {
			traceID = uuid.NewString()
		}
		requestID := firstHeader(md, "x-request-id")
		if requestID == "" {
			requestID = uuid.NewString()
		}

		ctx = context.WithValue(ctx, ctxKeyTraceID, traceID)
		ctx = context.WithValue(ctx, ctxKeyRequestID, requestID)

		_ = grpc.SetHeader(ctx, metadata.Pairs(
			"x-trace-id", traceID,
			"x-request-id", requestID,
		))

		return handler(ctx, req)
	}
}

func authInterceptor(cfg AuthConfig) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		if !cfg.Enabled {
			return handler(ctx, req)
		}
		md, _ := metadata.FromIncomingContext(ctx)
		token := firstHeader(md, "x-internal-token")
		if token == "" || token != cfg.Token {
			return nil, status.Error(codes.Unauthenticated, "invalid internal token")
		}
		return handler(ctx, req)
	}
}

func loggingInterceptor(logger *zap.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		start := time.Now()
		resp, err := handler(ctx, req)
		code := status.Code(err)
		latency := time.Since(start)

		logger.Info("grpc.request",
			zap.String("service", "user-service"),
			zap.String("method", info.FullMethod),
			zap.String("trace_id", TraceIDFromContext(ctx)),
			zap.String("request_id", RequestIDFromContext(ctx)),
			zap.String("status", code.String()),
			zap.Int64("latency_ms", latency.Milliseconds()),
		)

		return resp, err
	}
}

func recoveryInterceptor(logger *zap.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		defer func() {
			if r := recover(); r != nil {
				logger.Error("grpc.panic", zap.Any("panic", r), zap.String("method", info.FullMethod))
				err = status.Error(codes.Internal, "internal error")
			}
		}()
		return handler(ctx, req)
	}
}

func firstHeader(md metadata.MD, key string) string {
	if md == nil {
		return ""
	}
	vals := md.Get(key)
	if len(vals) == 0 {
		return ""
	}
	return vals[0]
}

func mapError(err error, msg string) error {
	if err == nil {
		return nil
	}
	return status.Error(codes.Internal, fmt.Sprintf("%s: %v", msg, err))
}
