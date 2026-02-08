package metrics

import (
	"sync/atomic"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type AuthMetrics struct {
	Requests atomic.Int64
	Errors   atomic.Int64
	prom     *PromMetrics
}

type PromMetrics struct {
	Requests *prometheus.CounterVec
	Latency  *prometheus.HistogramVec
}

func NewPromMetrics(service string) *PromMetrics {
	return &PromMetrics{
		Requests: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name:        "user_auth_requests_total",
				Help:        "Total auth requests",
				ConstLabels: prometheus.Labels{"service": service},
			},
			[]string{"endpoint", "status"},
		),
		Latency: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:        "user_auth_latency_seconds",
				Help:        "Auth request latency",
				ConstLabels: prometheus.Labels{"service": service},
				Buckets:     prometheus.DefBuckets,
			},
			[]string{"endpoint"},
		),
	}
}

func (m *AuthMetrics) AttachProm(pm *PromMetrics) {
	m.prom = pm
}

func (m *AuthMetrics) Record(endpoint string, status string, dur time.Duration) {
	if m == nil {
		return
	}
	m.Requests.Add(1)
	if status != "ok" {
		m.Errors.Add(1)
	}
	if m.prom != nil {
		m.prom.Requests.WithLabelValues(endpoint, status).Inc()
		m.prom.Latency.WithLabelValues(endpoint).Observe(dur.Seconds())
	}
}
