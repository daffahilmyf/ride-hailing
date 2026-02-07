package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/daffahilmyf/ride-hailing/services/gateway/internal/app/contextdata"
)

type fakeLimiter struct {
	allowed   bool
	remaining int
	reset     time.Time
	err       error
	key       string
}

func (f *fakeLimiter) Allow(key string) (bool, int, time.Time, error) {
	f.key = key
	return f.allowed, f.remaining, f.reset, f.err
}

func TestRateLimitMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		limiter    *fakeLimiter
		userID     string
		wantStatus int
	}{
		{
			name: "allowed",
			limiter: &fakeLimiter{
				allowed:   true,
				remaining: 9,
				reset:     time.Unix(1700000000, 0),
			},
			userID:     "user-1",
			wantStatus: http.StatusOK,
		},
		{
			name: "denied",
			limiter: &fakeLimiter{
				allowed:   false,
				remaining: 0,
			},
			userID:     "user-2",
			wantStatus: http.StatusTooManyRequests,
		},
		{
			name: "error",
			limiter: &fakeLimiter{
				allowed:   false,
				remaining: 0,
				err:       http.ErrHandlerTimeout,
			},
			userID:     "user-3",
			wantStatus: http.StatusTooManyRequests,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := gin.New()
			r.Use(func(c *gin.Context) {
				if tt.userID != "" {
					contextdata.SetUserContext(c, tt.userID, "rider")
				}
				c.Next()
			})
			r.Use(RateLimitMiddleware(tt.limiter, 10))
			r.GET("/ping", func(c *gin.Context) {
				c.Status(http.StatusOK)
			})

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/ping", nil)
			r.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Fatalf("expected %d, got %d", tt.wantStatus, w.Code)
			}
			if tt.userID != "" && tt.limiter.key != tt.userID {
				t.Fatalf("expected key %q, got %q", tt.userID, tt.limiter.key)
			}
			if got := w.Header().Get("X-RateLimit-Limit"); got == "" {
				t.Fatalf("expected X-RateLimit-Limit header")
			}
		})
	}
}
