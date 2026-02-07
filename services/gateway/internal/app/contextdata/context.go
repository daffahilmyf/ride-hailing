package contextdata

import "github.com/gin-gonic/gin"

type ctxKey string

const (
	ctxKeyTraceID   ctxKey = "trace_id"
	ctxKeyRequestID ctxKey = "request_id"
	ctxKeyUserID    ctxKey = "user_id"
	ctxKeyRole      ctxKey = "role"
)

func SetTraceID(c *gin.Context, traceID string) {
	if traceID != "" {
		c.Set(string(ctxKeyTraceID), traceID)
	}
}

func SetRequestID(c *gin.Context, requestID string) {
	if requestID != "" {
		c.Set(string(ctxKeyRequestID), requestID)
	}
}

func SetUserContext(c *gin.Context, userID, role string) {
	if userID != "" {
		c.Set(string(ctxKeyUserID), userID)
	}
	if role != "" {
		c.Set(string(ctxKeyRole), role)
	}
}

func GetTraceID(c *gin.Context) string {
	v, _ := c.Get(string(ctxKeyTraceID))
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

func GetRequestID(c *gin.Context) string {
	v, _ := c.Get(string(ctxKeyRequestID))
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

func GetUserID(c *gin.Context) string {
	v, _ := c.Get(string(ctxKeyUserID))
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

func GetRole(c *gin.Context) string {
	v, _ := c.Get(string(ctxKeyRole))
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}
