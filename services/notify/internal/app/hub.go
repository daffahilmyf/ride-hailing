package app

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"
)

type Notification struct {
	Event     string      `json:"event"`
	Subject   string      `json:"subject"`
	TraceID   string      `json:"trace_id,omitempty"`
	RequestID string      `json:"request_id,omitempty"`
	Data      interface{} `json:"data,omitempty"`
	Timestamp string      `json:"timestamp"`
}

type Filters struct {
	DriverID string
	RiderID  string
	RideID   string
}

type Subscriber struct {
	ID      string
	Filters Filters
	Ch      chan Notification
}

type Hub struct {
	mu          sync.RWMutex
	subscribers map[string]*Subscriber
	bufferSize  int
}

func NewHub(bufferSize int) *Hub {
	if bufferSize <= 0 {
		bufferSize = 64
	}
	return &Hub{
		subscribers: make(map[string]*Subscriber),
		bufferSize:  bufferSize,
	}
}

func (h *Hub) Subscribe(id string, filters Filters) *Subscriber {
	h.mu.Lock()
	defer h.mu.Unlock()
	s := &Subscriber{ID: id, Filters: filters, Ch: make(chan Notification, h.bufferSize)}
	h.subscribers[id] = s
	return s
}

func (h *Hub) Unsubscribe(id string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if sub, ok := h.subscribers[id]; ok {
		close(sub.Ch)
		delete(h.subscribers, id)
	}
}

func (h *Hub) Broadcast(n Notification, data map[string]interface{}) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for _, sub := range h.subscribers {
		if !matches(sub.Filters, data) {
			continue
		}
		select {
		case sub.Ch <- n:
		default:
			// drop if subscriber is slow
		}
	}
}

func matches(f Filters, data map[string]interface{}) bool {
	if f.DriverID == "" && f.RiderID == "" && f.RideID == "" {
		return true
	}
	if f.DriverID != "" {
		if v, ok := data["driver_id"].(string); ok && v == f.DriverID {
			return true
		}
	}
	if f.RiderID != "" {
		if v, ok := data["rider_id"].(string); ok && v == f.RiderID {
			return true
		}
	}
	if f.RideID != "" {
		if v, ok := data["ride_id"].(string); ok && v == f.RideID {
			return true
		}
	}
	return false
}

func WriteSSE(w http.ResponseWriter, n Notification) error {
	data, err := json.Marshal(n)
	if err != nil {
		return err
	}
	_, err = w.Write([]byte("event: " + n.Event + "\n"))
	if err != nil {
		return err
	}
	_, err = w.Write([]byte("data: " + string(data) + "\n\n"))
	return err
}

func KeepAlive(w http.ResponseWriter, interval time.Duration) {
	if interval <= 0 {
		return
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for range ticker.C {
		_, _ = w.Write([]byte(": keepalive\n\n"))
		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		}
	}
}
