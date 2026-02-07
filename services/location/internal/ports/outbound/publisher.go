package outbound

import "context"

type EventPublisher interface {
	Publish(ctx context.Context, subject string, payload []byte) error
}
