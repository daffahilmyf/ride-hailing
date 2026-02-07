package metrics

import "sync/atomic"

type OutboxMetrics struct {
	claimed   atomic.Int64
	published atomic.Int64
	failed    atomic.Int64
	dlq       atomic.Int64
}

func (m *OutboxMetrics) IncClaimed(n int) {
	if m == nil {
		return
	}
	m.claimed.Add(int64(n))
}

func (m *OutboxMetrics) IncPublished() {
	if m == nil {
		return
	}
	m.published.Add(1)
}

func (m *OutboxMetrics) IncFailed() {
	if m == nil {
		return
	}
	m.failed.Add(1)
}

func (m *OutboxMetrics) IncDLQ() {
	if m == nil {
		return
	}
	m.dlq.Add(1)
}
