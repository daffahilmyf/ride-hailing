package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/daffahilmyf/ride-hailing/services/gateway/internal/app/contextdata"
	"github.com/daffahilmyf/ride-hailing/services/gateway/internal/app/responses"
)

const (
	RoleRider   = "rider"
	RoleDriver  = "driver"
	scopesClaim = "scopes"
)

func RequireRole(allowed ...string) gin.HandlerFunc {
	allowedSet := map[string]struct{}{}
	for _, role := range allowed {
		allowedSet[role] = struct{}{}
	}

	return func(c *gin.Context) {
		role := contextdata.GetRole(c)
		if role == "" {
			responses.RespondErrorCode(c, responses.CodeForbidden, map[string]string{"reason": "MISSING_ROLE"})
			c.Abort()
			return
		}
		if _, ok := allowedSet[role]; !ok {
			responses.RespondErrorCode(c, responses.CodeForbidden, map[string]string{"reason": "ROLE_NOT_ALLOWED"})
			c.Abort()
			return
		}
		c.Next()
	}
}

func RequireScope(required ...string) gin.HandlerFunc {
	requiredSet := map[string]struct{}{}
	for _, s := range required {
		requiredSet[s] = struct{}{}
	}

	return func(c *gin.Context) {
		raw, ok := c.Get(scopesClaim)
		if !ok {
			responses.RespondErrorCode(c, responses.CodeForbidden, map[string]string{"reason": "MISSING_SCOPES"})
			c.Abort()
			return
		}
		scopes := splitScopes(raw)
		for _, r := range required {
			if _, ok := scopes[r]; !ok {
				responses.RespondErrorCode(c, responses.CodeForbidden, map[string]string{"reason": "SCOPE_NOT_ALLOWED"})
				c.Abort()
				return
			}
		}
		c.Next()
	}
}

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

func splitScopes(raw interface{}) map[string]struct{} {
	out := map[string]struct{}{}
	switch v := raw.(type) {
	case string:
		for _, s := range strings.Fields(v) {
			out[s] = struct{}{}
		}
	case []string:
		for _, s := range v {
			out[s] = struct{}{}
		}
	}
	return out
}
