package grpc

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

type Validator interface {
	Validate() error
}

func UnaryInterceptors(logger *zap.Logger, metrics *Metrics) []grpc.UnaryServerInterceptor {
	return []grpc.UnaryServerInterceptor{
		recoveryInterceptor(logger),
		requestIDInterceptor(),
		validationInterceptor(),
		metricsInterceptor(metrics),
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

func validationInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		if v, ok := req.(Validator); ok {
			if err := v.Validate(); err != nil {
				return nil, status.Error(codes.InvalidArgument, err.Error())
			}
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
			zap.String("service", "ride-service"),
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
				logger.Error("grpc.panic",
					zap.String("service", "ride-service"),
					zap.String("method", info.FullMethod),
					zap.String("trace_id", TraceIDFromContext(ctx)),
					zap.String("request_id", RequestIDFromContext(ctx)),
					zap.String("panic", fmt.Sprint(r)),
				)
				err = status.Error(codes.Internal, "internal server error")
			}
		}()
		return handler(ctx, req)
	}
}

func metricsInterceptor(metrics *Metrics) grpc.UnaryServerInterceptor {
	if metrics == nil {
		metrics = NewMetrics()
	}
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		start := time.Now()
		resp, err := handler(ctx, req)
		metrics.Record(info.FullMethod, status.Code(err), time.Since(start))
		return resp, err
	}
}

func firstHeader(md metadata.MD, key string) string {
	values := md.Get(key)
	if len(values) == 0 {
		return ""
	}
	return values[0]
}
