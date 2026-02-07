package middleware

import (
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"

	"github.com/gin-gonic/gin"
)

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
