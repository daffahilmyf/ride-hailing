package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"

	"github.com/daffahilmyf/ride-hailing/services/gateway/internal/app/contextdata"
)

func TestLoggerMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	core, logs := observer.New(zap.InfoLevel)
	logger := zap.New(core)

	r := gin.New()
	r.Use(LoggerMiddleware(logger, "api-gateway"))
	r.GET("/ping", func(c *gin.Context) {
		c.Header("X-User-ID", contextdata.GetUserID(c))
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/ping", nil)
	r.ServeHTTP(w, req)

	if w.Header().Get(HeaderTraceID) == "" {
		t.Fatalf("expected %s header", HeaderTraceID)
	}
	if w.Header().Get(HeaderRequestID) == "" {
		t.Fatalf("expected %s header", HeaderRequestID)
	}

	entries := logs.FilterMessage("http.request").All()
	if len(entries) != 1 {
		t.Fatalf("expected 1 log entry, got %d", len(entries))
	}
}
