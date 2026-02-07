package middleware

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func LogGeneralError(c *gin.Context, logger *zap.Logger, msg string, err error) {
	fields := []zap.Field{
		zap.String("trace_id", GetTraceID(c)),
		zap.String("request_id", GetRequestID(c)),
	}
	if err != nil {
		fields = append(fields, zap.Error(err))
	}
	logger.Warn(msg, fields...)
}
