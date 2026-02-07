package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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
		logger.Info("http.request",
			zap.String("service", serviceName),
			zap.String("trace_id", traceID),
			zap.String("request_id", requestID),
			zap.String("method", c.Request.Method),
			zap.String("path", c.FullPath()),
			zap.Int("status", c.Writer.Status()),
			zap.Int64("latency_ms", latency.Milliseconds()),
		)
	}
}
