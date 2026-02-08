package middleware

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"

	"github.com/daffahilmyf/ride-hailing/services/gateway/internal/app/contextdata"
)

const (
	HeaderTraceID   = "X-Trace-Id"
	HeaderRequestID = "X-Request-Id"
)

func LoggerMiddleware(logger *zap.Logger, serviceName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		traceID := c.GetHeader(HeaderTraceID)
		if traceID == "" {
			traceID = uuid.NewString()
			c.Header(HeaderTraceID, traceID)
		}

		requestID := c.GetHeader(HeaderRequestID)
		if requestID == "" {
			requestID = uuid.NewString()
			c.Header(HeaderRequestID, requestID)
		}

		contextdata.SetTraceID(c, traceID)
		contextdata.SetRequestID(c, requestID)

		c.Next()

		latency := time.Since(start)
		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path
		}
		logger.Info("http.request",
			zap.String("service", serviceName),
			zap.String("trace_id", traceID),
			zap.String("request_id", requestID),
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.Int("status", c.Writer.Status()),
			zap.Int64("latency_ms", latency.Milliseconds()),
		)
	}
}

func LogGeneralError(c *gin.Context, logger *zap.Logger, msg string, err error) {
	fields := []zap.Field{
		zap.String("trace_id", contextdata.GetTraceID(c)),
		zap.String("request_id", contextdata.GetRequestID(c)),
	}
	if err != nil {
		fields = append(fields, zap.Error(err))
	}
	logger.Warn(msg, fields...)
}

func RequestTimeout(timeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.URL.Path == "/v1/notify/sse" || strings.Contains(c.Request.Header.Get("Accept"), "text/event-stream") {
			c.Next()
			return
		}
		if timeout <= 0 {
			c.Next()
			return
		}
		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		defer cancel()
		c.Request = c.Request.WithContext(ctx)
		c.Next()
		if ctx.Err() == context.DeadlineExceeded {
			c.Writer.WriteHeader(http.StatusGatewayTimeout)
		}
	}
}

func MaxBodyBytes(n int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		if n > 0 {
			c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, n)
		}
		c.Next()
	}
}

func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("Referrer-Policy", "no-referrer")
		c.Header("X-XSS-Protection", "0")
		c.Next()
	}
}

func TraceMiddleware(serviceName string) gin.HandlerFunc {
	tracer := otel.Tracer(serviceName)
	return func(c *gin.Context) {
		ctx, span := tracer.Start(c.Request.Context(), c.FullPath())
		defer span.End()
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

func SpanFromContext(c *gin.Context) trace.Span {
	return trace.SpanFromContext(c.Request.Context())
}
