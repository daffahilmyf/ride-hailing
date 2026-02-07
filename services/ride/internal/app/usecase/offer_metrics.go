package usecase

import "sync/atomic"

type OfferMetrics struct {
	created  atomic.Int64
	accepted atomic.Int64
	declined atomic.Int64
	expired  atomic.Int64
}

func (m *OfferMetrics) IncCreated() {
	if m == nil {
		return
	}
	m.created.Add(1)
}

func (m *OfferMetrics) IncAccepted() {
	if m == nil {
		return
	}
	m.accepted.Add(1)
}

func (m *OfferMetrics) IncDeclined() {
	if m == nil {
		return
	}
	m.declined.Add(1)
}

func (m *OfferMetrics) IncExpired() {
	if m == nil {
		return
	}
	m.expired.Add(1)
}
