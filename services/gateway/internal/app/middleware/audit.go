package middleware

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/daffahilmyf/ride-hailing/services/gateway/internal/app/contextdata"
)

func AuditLogger(logger *zap.Logger, action string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		logger.Info("audit",
			zap.String("action", action),
			zap.String("user_id", contextdata.GetUserID(c)),
			zap.String("trace_id", contextdata.GetTraceID(c)),
			zap.String("request_id", contextdata.GetRequestID(c)),
		)
	}
}
