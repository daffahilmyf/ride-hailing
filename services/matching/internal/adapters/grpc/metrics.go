package grpc

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/grpc/codes"
)

type Metrics struct {
	total  atomic.Int64
	errors atomic.Int64
	mu     sync.Mutex
	byCode map[codes.Code]int64
	prom   *PromMetrics
}

func NewMetrics() *Metrics {
	return &Metrics{
		byCode: make(map[codes.Code]int64),
	}
}

func (m *Metrics) AttachProm(pm *PromMetrics) {
	m.prom = pm
}

func (m *Metrics) Record(method string, code codes.Code, latency time.Duration) {
	if m == nil {
		return
	}
	m.total.Add(1)
	if code != codes.OK {
		m.errors.Add(1)
	}
	m.mu.Lock()
	m.byCode[code]++
	m.mu.Unlock()

	if m.prom != nil {
		m.prom.Requests.WithLabelValues(code.String()).Inc()
		m.prom.Latency.WithLabelValues(method, code.String()).Observe(latency.Seconds())
	}
}

type PromMetrics struct {
	Requests *prometheus.CounterVec
	Latency  *prometheus.HistogramVec
}

func NewPromMetrics(service string) *PromMetrics {
	return &PromMetrics{
		Requests: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name:        "grpc_requests_total",
				Help:        "Total gRPC requests",
				ConstLabels: prometheus.Labels{"service": service},
			},
			[]string{"code"},
		),
		Latency: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:        "grpc_request_duration_seconds",
				Help:        "gRPC request latency in seconds",
				ConstLabels: prometheus.Labels{"service": service},
			},
			[]string{"method", "code"},
		),
	}
}
