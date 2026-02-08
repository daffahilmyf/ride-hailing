package workers

import (
	"context"

	"github.com/daffahilmyf/ride-hailing/services/notify/internal/adapters/broker"
	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
)

type EventConsumer struct {
	Consumer *broker.Consumer
	Subject  string
	Durable  string
	Batch    int
	Logger   *zap.Logger
	Handler  func(ctx context.Context, payload []byte) error
}

func (c *EventConsumer) Run(ctx context.Context) error {
	if c == nil || c.Consumer == nil || c.Handler == nil {
		return nil
	}
	return c.Consumer.Pull(ctx, c.Subject, c.Durable, c.Batch, func(msg *nats.Msg) error {
		if err := c.Handler(ctx, msg.Data); err != nil {
			if c.Logger != nil {
				c.Logger.Warn("event.handle_failed", zap.String("subject", c.Subject), zap.Error(err))
			}
			return err
		}
		return nil
	})
}
