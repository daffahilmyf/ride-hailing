package broker

import (
	"context"
	"time"

	"github.com/nats-io/nats.go"
)

type Publisher struct {
	js nats.JetStreamContext
}

func NewPublisher(js nats.JetStreamContext) *Publisher {
	return &Publisher{js: js}
}

func (p *Publisher) Publish(ctx context.Context, subject string, payload []byte) error {
	if p == nil || p.js == nil {
		return nil
	}
	_, err := p.js.PublishMsg(&nats.Msg{
		Subject: subject,
		Data:    payload,
	}, nats.Context(ctx))
	return err
}

func (p *Publisher) PublishWithTimeout(ctx context.Context, subject string, payload []byte, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	return p.Publish(ctx, subject, payload)
}
