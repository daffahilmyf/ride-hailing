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
	return EventEnvelope{
		ID:        uuid.NewString(),
		Type:      eventType,
		Source:    source,
		Time:      time.Now().UTC().Format(time.RFC3339Nano),
		Version:   "v1",
		TraceID:   traceID,
		RequestID: requestID,
		Payload:   payload,
	}
}
