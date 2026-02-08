package app

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"
)

type Notification struct {
	ID        string      `json:"id"`
	Event     string      `json:"event"`
	Subject   string      `json:"subject"`
	TraceID   string      `json:"trace_id,omitempty"`
	RequestID string      `json:"request_id,omitempty"`
	Data      interface{} `json:"data,omitempty"`
	Timestamp string      `json:"timestamp"`
	Schema    int         `json:"schema_version"`
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
	history     []Notification
	historySize int
	lastID      uint64
}

func NewHub(bufferSize, historySize int) *Hub {
	if bufferSize <= 0 {
		bufferSize = 64
	}
	if historySize < 0 {
		historySize = 0
	}
	return &Hub{
		subscribers: make(map[string]*Subscriber),
		bufferSize:  bufferSize,
		historySize: historySize,
	}
}

func (h *Hub) Subscribe(id string, filters Filters) *Subscriber {
	h.mu.Lock()
	defer h.mu.Unlock()
	s := &Subscriber{ID: id, Filters: filters, Ch: make(chan Notification, h.bufferSize)}
	h.subscribers[id] = s
	return s
}

func (h *Hub) SubscribeWithReplay(id string, filters Filters, lastEventID uint64) *Subscriber {
	h.mu.Lock()
	defer h.mu.Unlock()
	s := &Subscriber{ID: id, Filters: filters, Ch: make(chan Notification, h.bufferSize)}
	h.subscribers[id] = s
	if lastEventID > 0 && len(h.history) > 0 {
		for _, n := range h.history {
			if n.ID == "" {
				continue
			}
			if nID, err := ParseEventID(n.ID); err == nil && nID > lastEventID {
				if matches(filters, mapData(n.Data)) {
					select {
					case s.Ch <- n:
					default:
						return s
					}
				}
			}
		}
	}
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

func (h *Hub) Broadcast(n Notification, data map[string]interface{}) (int, int) {
	n.ID = h.nextID()
	n.Schema = 1
	h.appendHistory(n)
	h.mu.RLock()
	defer h.mu.RUnlock()
	sent := 0
	dropped := 0
	for _, sub := range h.subscribers {
		if !matches(sub.Filters, data) {
			continue
		}
		select {
		case sub.Ch <- n:
			sent++
		default:
			// drop if subscriber is slow
			dropped++
		}
	}
	return sent, dropped
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
	if n.ID != "" {
		if _, err = w.Write([]byte("id: " + n.ID + "\n")); err != nil {
			return err
		}
	}
	_, err = w.Write([]byte("event: " + n.Event + "\n"))
	if err != nil {
		return err
	}
	_, err = w.Write([]byte("data: " + string(data) + "\n\n"))
	return err
}

func (h *Hub) Count() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.subscribers)
}

func (h *Hub) nextID() string {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.lastID++
	return FormatEventID(h.lastID)
}

func (h *Hub) appendHistory(n Notification) {
	if h.historySize <= 0 {
		return
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	h.history = append(h.history, n)
	if len(h.history) > h.historySize {
		h.history = h.history[len(h.history)-h.historySize:]
	}
}

func mapData(v interface{}) map[string]interface{} {
	if v == nil {
		return nil
	}
	if data, ok := v.(map[string]interface{}); ok {
		return data
	}
	return nil
}

func KeepAlive(ctx context.Context, w http.ResponseWriter, interval time.Duration) {
	if interval <= 0 {
		return
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			_, _ = w.Write([]byte(": keepalive\n\n"))
			if f, ok := w.(http.Flusher); ok {
				f.Flush()
			}
		}
	}
}

func ParseEventID(id string) (uint64, error) {
	return strconv.ParseUint(id, 10, 64)
}

func FormatEventID(id uint64) string {
	return fmt.Sprintf("%d", id)
}
