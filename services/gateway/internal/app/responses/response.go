package responses

import (
	"github.com/gin-gonic/gin"

	"github.com/daffahilmyf/ride-hailing/services/gateway/internal/app/contextdata"
)

type APIError struct {
	Type    string      `json:"type"`
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

type APIResponse struct {
	Data  interface{}       `json:"data,omitempty"`
	Error *APIError         `json:"error,omitempty"`
	Meta  map[string]string `json:"meta,omitempty"`
}

func RespondOK(c *gin.Context, status int, data interface{}) {
	c.JSON(status, APIResponse{
		Data: data,
		Meta: buildMeta(c),
	})
}

func RespondError(c *gin.Context, status int, apiErr APIError) {
	c.JSON(status, APIResponse{
		Error: &apiErr,
		Meta:  buildMeta(c),
	})
}

func RespondErrorCode(c *gin.Context, code ErrorCode, details interface{}) {
	def := ErrorByCode(code)
	RespondError(c, def.HTTPStatus, APIError{
		Type:    def.Type,
		Code:    def.Code,
		Message: def.Message,
		Details: details,
	})
}

func RespondNotImplemented(c *gin.Context) {
	RespondErrorCode(c, CodeNotImplemented, nil)
}

func buildMeta(c *gin.Context) map[string]string {
	meta := map[string]string{}
	if traceID := contextdata.GetTraceID(c); traceID != "" {
		meta["trace_id"] = traceID
	}
	if requestID := contextdata.GetRequestID(c); requestID != "" {
		meta["request_id"] = requestID
	}
	return meta
}
