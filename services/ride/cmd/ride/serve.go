package main

import (
	"context"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/daffahilmyf/ride-hailing/services/ride/internal/adapters/db"
	grpcadapter "github.com/daffahilmyf/ride-hailing/services/ride/internal/adapters/grpc"
	"github.com/daffahilmyf/ride-hailing/services/ride/internal/app/handlers"
	"github.com/daffahilmyf/ride-hailing/services/ride/internal/app/usecase"
	"github.com/daffahilmyf/ride-hailing/services/ride/internal/infra"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start ride service",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := infra.LoadConfig()
		logger := infra.NewLogger()
		defer logger.Sync()

		pg, err := db.NewPostgres(context.Background(), cfg.PostgresDSN)
		if err != nil {
			logger.Fatal("db.connect_failed", zap.Error(err))
		}

		repo := db.NewRideRepo(pg.DB)
		idem := db.NewIdempotencyRepo(pg.DB)
		idemCleanup := db.NewIdempotencyCleanup(pg.DB)
		outbox := db.NewOutboxRepo(pg.DB)
		offers := db.NewRideOfferRepo(pg.DB)
		txMgr := db.NewTxManager(pg.DB)
		uc := &usecase.RideService{
			Repo:        repo,
			Idempotency: idem,
			TxManager:   txMgr,
			Outbox:      outbox,
			Offers:      offers,
		}

		metrics := grpcadapter.NewMetrics()
		srv := grpcadapter.NewServer(logger, handlers.Dependencies{Usecase: uc}, metrics)

		lis, err := net.Listen("tcp", cfg.GRPCAddr)
		if err != nil {
			logger.Fatal("grpc.listen_failed", zap.Error(err))
		}

		go func() {
			if err := srv.GRPC().Serve(lis); err != nil {
				logger.Fatal("grpc.serve_failed", zap.Error(err))
			}
		}()

		startIdempotencyCleanup(logger, idemCleanup, time.Duration(cfg.IdempotencyTTLSeconds)*time.Second)
		waitForShutdown(srv.GRPC(), cfg.ShutdownTimeoutSeconds, logger)
		return nil
	},
}

func startIdempotencyCleanup(logger *zap.Logger, cleaner *db.IdempotencyCleanup, ttl time.Duration) {
	if cleaner == nil || ttl <= 0 {
		return
	}
	ticker := time.NewTicker(time.Hour)
	go func() {
		for range ticker.C {
			cutoff := time.Now().UTC().Add(-ttl)
			deleted, err := cleaner.DeleteBefore(context.Background(), cutoff)
			if err != nil {
				logger.Warn("idempotency.cleanup_failed", zap.Error(err))
				continue
			}
			if deleted > 0 {
				logger.Info("idempotency.cleanup", zap.Int64("deleted", deleted))
			}
		}
	}()
}

func waitForShutdown(srv *grpc.Server, timeoutSeconds int, logger *zap.Logger) {
	signals := []os.Signal{syscall.SIGINT, syscall.SIGTERM}
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, signals...)

	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeoutSeconds)*time.Second)
	defer cancel()

	done := make(chan struct{})
	go func() {
		srv.GracefulStop()
		close(done)
	}()

	select {
	case <-done:
		logger.Info("grpc.shutdown_complete")
	case <-ctx.Done():
		srv.Stop()
		logger.Warn("grpc.shutdown_timeout")
	}
}
