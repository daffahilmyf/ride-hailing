package middleware

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
)

type Metrics struct {
	Requests *prometheus.CounterVec
	Latency  *prometheus.HistogramVec
}

func NewMetrics(service string) *Metrics {
	return &Metrics{
		Requests: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name:        "http_requests_total",
				Help:        "Total HTTP requests",
				ConstLabels: prometheus.Labels{"service": service},
			},
			[]string{"method", "path", "status"},
		),
		Latency: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:        "http_request_duration_seconds",
				Help:        "HTTP request latency",
				ConstLabels: prometheus.Labels{"service": service},
				Buckets:     prometheus.DefBuckets,
			},
			[]string{"method", "path"},
		),
	}
}

func MetricsMiddleware(m *Metrics) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		latency := time.Since(start).Seconds()
		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path
		}
		status := strconv.Itoa(c.Writer.Status())
		m.Requests.WithLabelValues(c.Request.Method, path, status).Inc()
		m.Latency.WithLabelValues(c.Request.Method, path).Observe(latency)
	}
}
