//go:build integration

package integration

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/daffahilmyf/ride-hailing/services/ride/internal/adapters/broker"
	"github.com/daffahilmyf/ride-hailing/services/ride/internal/adapters/db"
	"github.com/daffahilmyf/ride-hailing/services/ride/internal/app"
	"github.com/daffahilmyf/ride-hailing/services/ride/internal/ports/outbound"
	"go.uber.org/zap"
)

func TestOutboxWorkerPublishes(t *testing.T) {
	dsn := os.Getenv("TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("TEST_POSTGRES_DSN not set")
	}

	conn, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	require.NoError(t, err)

	outbox := db.NewOutboxRepo(conn)
	require.NoError(t, outbox.Enqueue(context.Background(), outbound.OutboxMessage{
		Topic:   "ride.test",
		Payload: "{\"ok\":true}",
	}))

	natsURL := os.Getenv("TEST_NATS_URL")
	if natsURL == "" {
		t.Skip("TEST_NATS_URL not set")
	}

	nc, err := nats.Connect(natsURL)
	require.NoError(t, err)
	js, err := nc.JetStream()
	require.NoError(t, err)

	logger, _ := zap.NewDevelopment()
	worker := &app.OutboxWorker{
		Repo:        outbox,
		Publisher:   broker.NewPublisher(js),
		Logger:      logger,
		Metrics:     &app.OutboxMetrics{},
		BatchSize:   10,
		MaxAttempts: 3,
		Interval:    100 * time.Millisecond,
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go worker.Run(ctx)

	sub, err := js.SubscribeSync("ride.test")
	require.NoError(t, err)
	_, err = sub.NextMsg(2 * time.Second)
	require.NoError(t, err)
}
