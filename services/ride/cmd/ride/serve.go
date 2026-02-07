package main

import (
	"context"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/daffahilmyf/ride-hailing/services/ride/internal/adapters/broker"
	"github.com/daffahilmyf/ride-hailing/services/ride/internal/adapters/db"
	grpcadapter "github.com/daffahilmyf/ride-hailing/services/ride/internal/adapters/grpc"
	"github.com/daffahilmyf/ride-hailing/services/ride/internal/app/handlers"
	"github.com/daffahilmyf/ride-hailing/services/ride/internal/app/metrics"
	"github.com/daffahilmyf/ride-hailing/services/ride/internal/app/usecase"
	"github.com/daffahilmyf/ride-hailing/services/ride/internal/app/workers"
	"github.com/daffahilmyf/ride-hailing/services/ride/internal/infra"
	"github.com/nats-io/nats.go"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	health "google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
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
			Repo:         repo,
			Idempotency:  idem,
			TxManager:    txMgr,
			Outbox:       outbox,
			Offers:       offers,
			OfferMetrics: &usecase.OfferMetrics{},
		}

		grpcMetrics := grpcadapter.NewMetrics()
		srv := grpcadapter.NewServer(logger, handlers.Dependencies{Usecase: uc}, grpcMetrics, grpcadapter.AuthConfig{
			Enabled: cfg.InternalAuthEnabled,
			Token:   cfg.InternalAuthToken,
		})
		healthSrv := health.NewServer()
		healthpb.RegisterHealthServer(srv.GRPC(), healthSrv)
		reflection.Register(srv.GRPC())

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		if cfg.OutboxEnabled {
			nc, err := nats.Connect(cfg.NATSURL)
			if err != nil {
				logger.Fatal("nats.connect_failed", zap.Error(err))
			}
			logger.Info("nats.connected", zap.String("url", cfg.NATSURL))
			js, err := nc.JetStream()
			if err != nil {
				logger.Fatal("nats.jetstream_failed", zap.Error(err))
			}
			logger.Info("nats.jetstream_ready")
			ensureStream(logger, js, "RIDES", []string{"ride.*"})
			publisher := broker.NewPublisher(js)
			outboxMetrics := &metrics.OutboxMetrics{}
			worker := &workers.OutboxWorker{
				Repo:        outbox,
				Publisher:   publisher,
				Logger:      logger,
				Metrics:     outboxMetrics,
				BatchSize:   cfg.OutboxBatchSize,
				MaxAttempts: cfg.OutboxMaxAttempts,
				Interval:    time.Duration(cfg.OutboxIntervalMillis) * time.Millisecond,
			}
			go worker.Run(ctx)

			reporter := &workers.MetricsReporter{
				Outbox:   outboxMetrics,
				Logger:   logger,
				Interval: 30 * time.Second,
			}
			go reporter.Run(ctx)

			retention := time.Duration(cfg.OutboxRetentionHours) * time.Hour
			cleanup := &workers.OutboxCleanupWorker{
				Repo:      outbox,
				Logger:    logger,
				Retention: retention,
				Interval:  24 * time.Hour,
			}
			go cleanup.Run(ctx)
		}

		if cfg.OfferExpiryEnabled {
			expiry := &workers.OfferExpiryWorker{
				Repo:     offers,
				Usecase:  uc,
				Logger:   logger,
				Interval: time.Duration(cfg.OfferExpiryIntervalMs) * time.Millisecond,
				Batch:    cfg.OfferExpiryBatchSize,
			}
			go expiry.Run(ctx)
		}

		lis, err := net.Listen("tcp", cfg.GRPCAddr)
		if err != nil {
			logger.Fatal("grpc.listen_failed", zap.Error(err))
		}

		go func() {
			if err := srv.GRPC().Serve(lis); err != nil {
				logger.Fatal("grpc.serve_failed", zap.Error(err))
			}
		}()

		logger.Info("service.started",
			zap.String("service", cfg.ServiceName),
			zap.String("grpc_addr", cfg.GRPCAddr),
		)

		healthSrv.SetServingStatus("ride.v1.RideService", healthpb.HealthCheckResponse_SERVING)
		startIdempotencyCleanup(logger, idemCleanup, time.Duration(cfg.IdempotencyTTLSeconds)*time.Second)
		waitForShutdown(srv.GRPC(), cfg.ShutdownTimeoutSeconds, logger, cancel)
		return nil
	},
}

func ensureStream(logger *zap.Logger, js nats.JetStreamContext, name string, subjects []string) {
	if js == nil {
		return
	}
	if _, err := js.StreamInfo(name); err == nil {
		return
	}
	_, err := js.AddStream(&nats.StreamConfig{
		Name:      name,
		Subjects:  subjects,
		Storage:   nats.FileStorage,
		Retention: nats.LimitsPolicy,
	})
	if err != nil {
		logger.Warn("nats.stream_create_failed", zap.String("stream", name), zap.Error(err))
		return
	}
	logger.Info("nats.stream_created", zap.String("stream", name))
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

func waitForShutdown(srv *grpc.Server, timeoutSeconds int, logger *zap.Logger, cancel context.CancelFunc) {
	signals := []os.Signal{syscall.SIGINT, syscall.SIGTERM}
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, signals...)

	<-stop
	cancel()

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
