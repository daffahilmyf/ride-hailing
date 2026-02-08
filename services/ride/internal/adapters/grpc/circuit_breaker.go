package grpc

import (
	"errors"
	"sync"
	"time"
)

var ErrCircuitOpen = errors.New("circuit breaker open")

type State int

const (
	StateClosed State = iota
	StateOpen
	StateHalfOpen
)

func (s State) String() string {
	switch s {
	case StateClosed:
		return "closed"
	case StateOpen:
		return "open"
	case StateHalfOpen:
		return "half_open"
	default:
		return "unknown"
	}
}

type CircuitBreakerSettings struct {
	Name          string
	MaxRequests   uint32
	Interval      time.Duration
	Timeout       time.Duration
	FailureRatio  float64
	MinRequests   uint32
	OnStateChange func(name string, from State, to State)
}

type Counts struct {
	Requests      uint32
	TotalFailures uint32
}

type CircuitBreaker struct {
	settings         CircuitBreakerSettings
	state            State
	counts           Counts
	openedAt         time.Time
	intervalStarted  time.Time
	halfOpenRequests uint32
	halfOpenSuccess  uint32
	mu               sync.Mutex
}

func NewCircuitBreaker(settings CircuitBreakerSettings) *CircuitBreaker {
	return &CircuitBreaker{
		settings:        settings,
		state:           StateClosed,
		intervalStarted: time.Now(),
	}
}

func (cb *CircuitBreaker) Execute(fn func() (any, error)) (any, error) {
	if err := cb.beforeRequest(); err != nil {
		return nil, err
	}
	res, err := fn()
	cb.afterRequest(err)
	return res, err
}

func (cb *CircuitBreaker) beforeRequest() error {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	now := time.Now()
	if cb.state == StateOpen {
		if cb.settings.Timeout > 0 && now.Sub(cb.openedAt) >= cb.settings.Timeout {
			cb.setState(StateHalfOpen, now)
		} else {
			return ErrCircuitOpen
		}
	}

	if cb.state == StateHalfOpen {
		maxRequests := cb.settings.MaxRequests
		if maxRequests == 0 {
			maxRequests = 1
		}
		if cb.halfOpenRequests >= maxRequests {
			return ErrCircuitOpen
		}
		cb.halfOpenRequests++
		return nil
	}

	if cb.settings.Interval > 0 && now.Sub(cb.intervalStarted) >= cb.settings.Interval {
		cb.resetCounts(now)
	}

	return nil
}

func (cb *CircuitBreaker) afterRequest(err error) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.counts.Requests++
	if err != nil {
		cb.counts.TotalFailures++
	}

	switch cb.state {
	case StateHalfOpen:
		if err != nil {
			cb.openedAt = time.Now()
			cb.setState(StateOpen, cb.openedAt)
			return
		}
		cb.halfOpenSuccess++
		maxRequests := cb.settings.MaxRequests
		if maxRequests == 0 {
			maxRequests = 1
		}
		if cb.halfOpenSuccess >= maxRequests {
			cb.setState(StateClosed, time.Now())
			cb.resetCounts(time.Now())
		}
	case StateClosed:
		if cb.shouldTrip() {
			cb.openedAt = time.Now()
			cb.setState(StateOpen, cb.openedAt)
		}
	}
}

func (cb *CircuitBreaker) shouldTrip() bool {
	if cb.counts.Requests == 0 || cb.counts.Requests < cb.settings.MinRequests {
		return false
	}
	if cb.settings.FailureRatio <= 0 {
		return false
	}
	failureRatio := float64(cb.counts.TotalFailures) / float64(cb.counts.Requests)
	return failureRatio >= cb.settings.FailureRatio
}

func (cb *CircuitBreaker) setState(state State, now time.Time) {
	if cb.state == state {
		return
	}
	from := cb.state
	cb.state = state
	if state == StateHalfOpen {
		cb.halfOpenRequests = 0
		cb.halfOpenSuccess = 0
	}
	if state == StateClosed {
		cb.resetCounts(now)
	}
	if cb.settings.OnStateChange != nil {
		cb.settings.OnStateChange(cb.settings.Name, from, state)
	}
}

func (cb *CircuitBreaker) resetCounts(now time.Time) {
	cb.counts = Counts{}
	cb.intervalStarted = now
}
