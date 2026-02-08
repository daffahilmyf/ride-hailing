package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type OutboxEvent struct {
	ID        string
	Topic     string
	Payload   []byte
	CreatedAt time.Time
}

func NewOutboxEvent(topic string, payload any) (OutboxEvent, error) {
	return NewOutboxEventWith(topic, payload, time.Now().UTC(), uuid.NewString())
}

func NewOutboxEventWith(topic string, payload any, now time.Time, id string) (OutboxEvent, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return OutboxEvent{}, err
	}
	return OutboxEvent{
		ID:        id,
		Topic:     topic,
		Payload:   data,
		CreatedAt: now.UTC(),
	}, nil
}
