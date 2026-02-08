package domain

import (
	"time"

	"github.com/google/uuid"
)

type EventEnvelope struct {
	ID        string `json:"id"`
	Type      string `json:"type"`
	Source    string `json:"source"`
	Time      string `json:"time"`
	Version   string `json:"version"`
	TraceID   string `json:"trace_id"`
	RequestID string `json:"request_id"`
	Payload   any    `json:"payload"`
}

func NewEventEnvelope(eventType string, source string, traceID string, requestID string, payload any) EventEnvelope {
	return NewEventEnvelopeWith(eventType, source, traceID, requestID, payload, time.Now().UTC(), uuid.NewString())
}

func NewEventEnvelopeWith(eventType string, source string, traceID string, requestID string, payload any, now time.Time, id string) EventEnvelope {
	return EventEnvelope{
		ID:        id,
		Type:      eventType,
		Source:    source,
		Time:      now.UTC().Format(time.RFC3339Nano),
		Version:   "v1",
		TraceID:   traceID,
		RequestID: requestID,
		Payload:   payload,
	}
}
