package middleware

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/daffahilmyf/ride-hailing/services/gateway/internal/app/contextdata"
	"github.com/daffahilmyf/ride-hailing/services/gateway/internal/app/responses"
)

type RateLimiter interface {
	Allow(key string) (allowed bool, remaining int, reset time.Time, err error)
}

func RateLimitMiddleware(limiter RateLimiter, limit int) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := contextdata.GetUserID(c)
		if key == "" {
			key = c.ClientIP()
		}
		allowed, remaining, reset, err := limiter.Allow(key)
		c.Header("X-RateLimit-Limit", itoa(limit))
		c.Header("X-RateLimit-Remaining", itoa(remaining))
		if !reset.IsZero() {
			c.Header("X-RateLimit-Reset", itoa(int(reset.Unix())))
		}
		if err != nil || !allowed {
			responses.RespondErrorCode(c, responses.CodeRateLimited, nil)
			c.Abort()
			return
		}
		c.Next()
	}
}

func itoa(v int) string {
	return fmt.Sprintf("%d", v)
}
