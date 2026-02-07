package outbound

import "context"

type OutboxMessage struct {
	ID      string
	Topic   string
	Payload string
}

type OutboxRepo interface {
	Enqueue(ctx context.Context, msg OutboxMessage) error
}
