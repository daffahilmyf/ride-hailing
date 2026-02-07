package grpc

import (
	"sync"
	"sync/atomic"
	"time"

	"google.golang.org/grpc/codes"
)

type Metrics struct {
	total  atomic.Int64
	errors atomic.Int64
	mu     sync.Mutex
	byCode map[codes.Code]int64
}

func NewMetrics() *Metrics {
	return &Metrics{
		byCode: make(map[codes.Code]int64),
	}
}

func (m *Metrics) Record(_ string, code codes.Code, _ time.Duration) {
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
}
