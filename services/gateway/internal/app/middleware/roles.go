package middleware

import (
	"github.com/gin-gonic/gin"

	"github.com/daffahilmyf/ride-hailing/services/gateway/internal/app/contextdata"
	"github.com/daffahilmyf/ride-hailing/services/gateway/internal/app/responses"
)

const (
	RoleRider  = "rider"
	RoleDriver = "driver"
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
