package main

import (
	"context"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/daffahilmyf/ride-hailing/services/matching/internal/adapters/broker"
	grpcadapter "github.com/daffahilmyf/ride-hailing/services/matching/internal/adapters/grpc"
	redisadapter "github.com/daffahilmyf/ride-hailing/services/matching/internal/adapters/redis"
	"github.com/daffahilmyf/ride-hailing/services/matching/internal/app/handlers"
	"github.com/daffahilmyf/ride-hailing/services/matching/internal/app/metrics"
	"github.com/daffahilmyf/ride-hailing/services/matching/internal/app/usecase"
	"github.com/daffahilmyf/ride-hailing/services/matching/internal/app/workers"
	"github.com/daffahilmyf/ride-hailing/services/matching/internal/infra"
	"github.com/nats-io/nats.go"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	health "google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start matching service",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := infra.LoadConfig()
		logger := infra.NewLogger()
		defer logger.Sync()

		telemetryShutdown, err := infra.SetupTelemetry(context.Background(), cfg)
		if err != nil {
			logger.Fatal("telemetry.init_failed", zap.Error(err))
		}
		defer func() {
			_ = telemetryShutdown(context.Background())
		}()

		redisClient := redis.NewClient(&redis.Options{
			Addr:     cfg.RedisAddr,
			Password: cfg.RedisPassword,
			DB:       cfg.RedisDB,
		})
		if err := redisClient.Ping(context.Background()).Err(); err != nil {
			logger.Fatal("redis.connect_failed", zap.Error(err))
		}
		defer redisClient.Close()

		repo := redisadapter.NewDriverRepo(redisClient, cfg.GeoKey, cfg.StatusKey, cfg.AvailableKey, cfg.OfferKeyPrefix)

		conn, rideClient, err := grpcadapter.DialRideClient(cfg.RideServiceAddr)
		if err != nil {
			logger.Fatal("ride.connect_failed", zap.Error(err))
		}
		defer conn.Close()

		uc := &usecase.MatchingService{
			Repo:            repo,
			RideClient:      rideClient,
			OfferTTLSeconds: cfg.OfferTTLSeconds,
			MatchRadius:     cfg.MatchRadiusMeters,
			MatchLimit:      cfg.MatchLimit,
			InternalToken:   cfg.RideServiceToken,
			OfferRetryMax:   cfg.OfferRetryMax,
			OfferBackoffMs:  cfg.OfferRetryBackoffMs,
			OfferMaxBackoff: cfg.OfferRetryMaxBackoffMs,
		}

		grpcMetrics := grpcadapter.NewMetrics()
		if cfg.Observability.MetricsEnabled {
			promMetrics := grpcadapter.NewPromMetrics(cfg.ServiceName)
			matchMetrics := &metrics.MatchingMetrics{}
			matchProm := metrics.NewPromMetrics(cfg.ServiceName)
			matchMetrics.AttachProm(matchProm)
			uc.Metrics = matchMetrics
			registry := prometheus.NewRegistry()
			registry.MustRegister(
				promMetrics.Requests,
				promMetrics.Latency,
				matchProm.OffersSent,
				matchProm.OffersFailed,
				matchProm.OffersSkipped,
				matchProm.NoCandidates,
			)
			grpcMetrics.AttachProm(promMetrics)
			go serveMetrics(cfg.Observability.MetricsAddr, registry, logger)
		}
		srv := grpcadapter.NewServer(logger, handlers.Dependencies{Usecase: uc}, grpcMetrics, grpcadapter.AuthConfig{
			Enabled: cfg.InternalAuthEnabled,
			Token:   cfg.InternalAuthToken,
		})
		healthSrv := health.NewServer()
		healthpb.RegisterHealthServer(srv.GRPC(), healthSrv)
		reflection.Register(srv.GRPC())

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		if cfg.EventsEnabled {
			nc, err := nats.Connect(cfg.NATSURL)
			if err != nil {
				logger.Fatal("nats.connect_failed", zap.Error(err))
			}
			logger.Info("nats.connected", zap.String("url", cfg.NATSURL))
			defer nc.Close()
			js, err := nc.JetStream()
			if err != nil {
				logger.Fatal("nats.jetstream_failed", zap.Error(err))
			}
			logger.Info("nats.jetstream_ready")
			ensureStream(logger, js, "RIDES", []string{"ride.*"})
			ensureStream(logger, js, "DRIVERS", []string{"driver.*"})
			consumer := broker.NewConsumer(js)

			rideConsumer := &workers.EventConsumer{
				Consumer: consumer,
				Subject:  cfg.RideRequestedSubject,
				Durable:  "matching-ride-requested",
				Batch:    10,
				Logger:   logger,
				Handler:  uc.HandleRideRequested,
			}
			go func() {
				if err := rideConsumer.Run(ctx); err != nil {
					logger.Warn("event.consumer_stopped", zap.String("subject", cfg.RideRequestedSubject), zap.Error(err))
				}
			}()

			locationConsumer := &workers.EventConsumer{
				Consumer: consumer,
				Subject:  cfg.DriverLocationSubject,
				Durable:  "matching-driver-location",
				Batch:    50,
				Logger:   logger,
				Handler:  uc.HandleDriverLocation,
			}
			go func() {
				if err := locationConsumer.Run(ctx); err != nil {
					logger.Warn("event.consumer_stopped", zap.String("subject", cfg.DriverLocationSubject), zap.Error(err))
				}
			}()
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

		healthSrv.SetServingStatus("matching.v1.MatchingService", healthpb.HealthCheckResponse_SERVING)
		waitForShutdown(srv.GRPC(), cfg.ShutdownTimeoutSeconds, logger, cancel)
		return nil
	},
}

func waitForShutdown(srv *grpc.Server, timeoutSeconds int, logger *zap.Logger, cancel context.CancelFunc) {
	signals := []os.Signal{syscall.SIGINT, syscall.SIGTERM}
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, signals...)

	<-stop
	cancel()

	ctx, cancelTimeout := context.WithTimeout(context.Background(), time.Duration(timeoutSeconds)*time.Second)
	defer cancelTimeout()

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

func serveMetrics(addr string, registry *prometheus.Registry, logger *zap.Logger) {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))
	server := &http.Server{
		Addr:              addr,
		ReadHeaderTimeout: 3 * time.Second,
		Handler:           mux,
	}
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Warn("metrics.listen_failed", zap.Error(err))
	}
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
