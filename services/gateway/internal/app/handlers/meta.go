package handlers

import (
	"github.com/gin-gonic/gin"

	"github.com/daffahilmyf/ride-hailing/services/gateway/internal/app/contextdata"
	"github.com/daffahilmyf/ride-hailing/services/gateway/internal/app/responses"
)

func WithGRPCMeta(c *gin.Context, upstream string) *gin.Context {
	return responses.WithMeta(c, map[string]string{
		"trace_id":   contextdata.GetTraceID(c),
		"request_id": contextdata.GetRequestID(c),
		"upstream":   upstream,
	})
}
