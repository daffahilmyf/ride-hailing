package broker

import (
	"context"
	"time"

	"github.com/nats-io/nats.go"
)

type Consumer struct {
	js nats.JetStreamContext
}

func NewConsumer(js nats.JetStreamContext) *Consumer {
	return &Consumer{js: js}
}

func (c *Consumer) Pull(ctx context.Context, subject string, durable string, batch int, handler func(*nats.Msg) error) error {
	if c == nil || c.js == nil {
		return nil
	}
	sub, err := c.js.PullSubscribe(subject, durable)
	if err != nil {
		return err
	}
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		msgs, err := sub.Fetch(batch, nats.MaxWait(2*time.Second))
		if err != nil && err != nats.ErrTimeout {
			return err
		}
		for _, msg := range msgs {
			if err := handler(msg); err != nil {
				_ = msg.Nak()
				continue
			}
			_ = msg.Ack()
		}
	}
}
