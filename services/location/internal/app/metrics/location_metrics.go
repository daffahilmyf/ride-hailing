package metrics

import (
	"sync/atomic"

	"github.com/prometheus/client_golang/prometheus"
)

type LocationMetrics struct {
	StaleGeoRemoved atomic.Int64
	prom            *PromMetrics
}

type PromMetrics struct {
	StaleGeoRemoved prometheus.Counter
}

func NewPromMetrics(service string) *PromMetrics {
	return &PromMetrics{
		StaleGeoRemoved: prometheus.NewCounter(prometheus.CounterOpts{
			Name:        "location_geo_stale_removed_total",
			Help:        "Total number of stale drivers removed from geo index",
			ConstLabels: prometheus.Labels{"service": service},
		}),
	}
}

func (m *LocationMetrics) AttachProm(pm *PromMetrics) {
	m.prom = pm
}

func (m *LocationMetrics) IncStaleGeoRemoved(n int) {
	if m == nil || n <= 0 {
		return
	}
	m.StaleGeoRemoved.Add(int64(n))
	if m.prom != nil {
		m.prom.StaleGeoRemoved.Add(float64(n))
	}
}
