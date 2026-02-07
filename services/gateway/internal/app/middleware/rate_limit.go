package middleware

import (
	"github.com/gin-gonic/gin"

	"github.com/daffahilmyf/ride-hailing/services/gateway/internal/app/contextdata"
	"github.com/daffahilmyf/ride-hailing/services/gateway/internal/app/responses"
)

type RateLimiter interface {
	Allow(key string) (bool, error)
}

func RateLimitMiddleware(limiter RateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := contextdata.GetUserID(c)
		if key == "" {
			key = c.ClientIP()
		}
		ok, err := limiter.Allow(key)
		if err != nil || !ok {
			responses.RespondErrorCode(c, responses.CodeRateLimited, nil)
			c.Abort()
			return
		}
		c.Next()
	}
}
