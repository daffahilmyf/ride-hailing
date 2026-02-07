package metrics

import (
	"sync/atomic"

	"github.com/prometheus/client_golang/prometheus"
)

type MatchingMetrics struct {
	OffersSent    atomic.Int64
	OffersFailed  atomic.Int64
	OffersSkipped atomic.Int64
	NoCandidates  atomic.Int64
	prom          *PromMetrics
}

type PromMetrics struct {
	OffersSent    prometheus.Counter
	OffersFailed  prometheus.Counter
	OffersSkipped prometheus.Counter
	NoCandidates  prometheus.Counter
}

func NewPromMetrics(service string) *PromMetrics {
	return &PromMetrics{
		OffersSent: prometheus.NewCounter(prometheus.CounterOpts{
			Name:        "matching_offers_sent_total",
			Help:        "Total number of offers sent",
			ConstLabels: prometheus.Labels{"service": service},
		}),
		OffersFailed: prometheus.NewCounter(prometheus.CounterOpts{
			Name:        "matching_offers_failed_total",
			Help:        "Total number of offer attempts failed",
			ConstLabels: prometheus.Labels{"service": service},
		}),
		OffersSkipped: prometheus.NewCounter(prometheus.CounterOpts{
			Name:        "matching_offers_skipped_total",
			Help:        "Total number of offers skipped due to existing offer",
			ConstLabels: prometheus.Labels{"service": service},
		}),
		NoCandidates: prometheus.NewCounter(prometheus.CounterOpts{
			Name:        "matching_no_candidates_total",
			Help:        "Total number of rides with no candidates",
			ConstLabels: prometheus.Labels{"service": service},
		}),
	}
}

func (m *MatchingMetrics) AttachProm(pm *PromMetrics) {
	m.prom = pm
}

func (m *MatchingMetrics) IncSent() {
	if m == nil {
		return
	}
	m.OffersSent.Add(1)
	if m.prom != nil {
		m.prom.OffersSent.Inc()
	}
}

func (m *MatchingMetrics) IncFailed() {
	if m == nil {
		return
	}
	m.OffersFailed.Add(1)
	if m.prom != nil {
		m.prom.OffersFailed.Inc()
	}
}

func (m *MatchingMetrics) IncSkipped() {
	if m == nil {
		return
	}
	m.OffersSkipped.Add(1)
	if m.prom != nil {
		m.prom.OffersSkipped.Inc()
	}
}

func (m *MatchingMetrics) IncNoCandidates() {
	if m == nil {
		return
	}
	m.NoCandidates.Add(1)
	if m.prom != nil {
		m.prom.NoCandidates.Inc()
	}
}
