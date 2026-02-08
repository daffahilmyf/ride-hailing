package handlers

import (
	"context"
	"net"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type RedisLimiter interface {
	Allow(ctx context.Context, key string, limit int, window time.Duration) (bool, error)
}

type RateLimiter struct {
	Redis  RedisLimiter
	Limit  int
	Window time.Duration
	Prefix string
	OnLimit func(endpoint string)
}

func RateLimitMiddleware(limiter *RateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		if limiter == nil || limiter.Redis == nil {
			c.Next()
			return
		}
		key := clientKey(c)
		allowed, err := limiter.Redis.Allow(c.Request.Context(), limiter.Prefix+key, limiter.Limit, limiter.Window)
		if err != nil || !allowed {
			if limiter.OnLimit != nil {
				path := c.FullPath()
				if path == "" {
					path = c.Request.URL.Path
				}
				limiter.OnLimit(path)
			}
			c.AbortWithStatusJSON(429, gin.H{"error": "rate_limited"})
			return
		}
		c.Next()
	}
}

func InternalAuthMiddleware(enabled bool, token string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !enabled {
			c.Next()
			return
		}
		header := c.GetHeader("X-Internal-Token")
		if header == "" || header != token {
			c.AbortWithStatusJSON(401, gin.H{"error": "unauthorized"})
			return
		}
		c.Next()
	}
}

func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		traceID := c.GetHeader("X-Trace-Id")
		if traceID == "" {
			traceID = uuid.NewString()
		}
		requestID := c.GetHeader("X-Request-Id")
		if requestID == "" {
			requestID = uuid.NewString()
		}
		c.Set("trace_id", traceID)
		c.Set("request_id", requestID)
		c.Writer.Header().Set("X-Trace-Id", traceID)
		c.Writer.Header().Set("X-Request-Id", requestID)
		c.Next()
	}
}

func GetTraceID(c *gin.Context) string {
	if val, ok := c.Get("trace_id"); ok {
		if s, ok := val.(string); ok {
			return s
		}
	}
	return ""
}

func GetRequestID(c *gin.Context) string {
	if val, ok := c.Get("request_id"); ok {
		if s, ok := val.(string); ok {
			return s
		}
	}
	return ""
}

func clientKey(c *gin.Context) string {
	ip := c.ClientIP()
	if ip == "" {
		ip = c.Request.RemoteAddr
	}
	host, _, err := net.SplitHostPort(ip)
	if err == nil && host != "" {
		ip = host
	}
	path := c.FullPath()
	if path == "" {
		path = c.Request.URL.Path
	}
	return strings.Join([]string{ip, path}, ":")
}
