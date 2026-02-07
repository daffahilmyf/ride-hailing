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
	data, err := json.Marshal(payload)
	if err != nil {
		return OutboxEvent{}, err
	}
	return OutboxEvent{
		ID:        uuid.NewString(),
		Topic:     topic,
		Payload:   data,
		CreatedAt: time.Now().UTC(),
	}, nil
}
