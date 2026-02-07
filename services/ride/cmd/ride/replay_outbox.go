package main

import (
	"context"

	"github.com/daffahilmyf/ride-hailing/services/ride/internal/adapters/db"
	"github.com/daffahilmyf/ride-hailing/services/ride/internal/infra"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var replayOutboxCmd = &cobra.Command{
	Use:   "outbox-replay",
	Short: "Replay FAILED outbox messages",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := infra.LoadConfig()
		logger := infra.NewLogger()
		defer logger.Sync()

		pg, err := db.NewPostgres(context.Background(), cfg.PostgresDSN)
		if err != nil {
			logger.Fatal("db.connect_failed", zap.Error(err))
		}

		outbox := db.NewOutboxRepo(pg.DB)
		count, err := outbox.ResetFailed(context.Background(), 100)
		if err != nil {
			logger.Fatal("outbox.replay_failed", zap.Error(err))
		}
		logger.Info("outbox.replay", zap.Int64("reset", count))
		return nil
	},
}
