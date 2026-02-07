package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/daffahilmyf/ride-hailing/services/gateway/internal/app/responses"
)

const scopesClaim = "scopes"

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
