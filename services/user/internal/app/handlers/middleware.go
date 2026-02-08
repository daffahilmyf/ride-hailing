package handlers

import (
	"net"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type RateLimiter struct {
	mu     sync.Mutex
	limits map[string]*rateEntry
	limit  int
	window time.Duration
}

type rateEntry struct {
	count     int
	resetTime time.Time
}

func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		limits: map[string]*rateEntry{},
		limit:  limit,
		window: window,
	}
}

func (r *RateLimiter) Allow(key string) bool {
	if r == nil || r.limit <= 0 {
		return true
	}
	if key == "" {
		return false
	}
	now := time.Now()
	r.mu.Lock()
	defer r.mu.Unlock()
	entry, ok := r.limits[key]
	if !ok || now.After(entry.resetTime) {
		r.limits[key] = &rateEntry{count: 1, resetTime: now.Add(r.window)}
		return true
	}
	if entry.count >= r.limit {
		return false
	}
	entry.count++
	return true
}

func RateLimitMiddleware(limiter *RateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := clientKey(c)
		if !limiter.Allow(key) {
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
