package domain

import (
	"time"

	"github.com/google/uuid"
)

type EventEnvelope struct {
	ID         string    `json:"id"`
	Type       string    `json:"type"`
	Version    string    `json:"version"`
	OccurredAt time.Time `json:"occurred_at"`
	Producer   string    `json:"producer"`
	TraceID    string    `json:"trace_id,omitempty"`
	RequestID  string    `json:"request_id,omitempty"`
	Data       any       `json:"data"`
}

func NewEventEnvelope(eventType string, producer string, traceID string, requestID string, data any) EventEnvelope {
	return EventEnvelope{
		ID:         uuid.NewString(),
		Type:       eventType,
		Version:    "v1",
		OccurredAt: time.Now().UTC(),
		Producer:   producer,
		TraceID:    traceID,
		RequestID:  requestID,
		Data:       data,
	}
}
