package main

import (
	"context"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/daffahilmyf/ride-hailing/services/location/internal/adapters/broker"
	grpcadapter "github.com/daffahilmyf/ride-hailing/services/location/internal/adapters/grpc"
	redisadapter "github.com/daffahilmyf/ride-hailing/services/location/internal/adapters/redis"
	"github.com/daffahilmyf/ride-hailing/services/location/internal/app/handlers"
	"github.com/daffahilmyf/ride-hailing/services/location/internal/app/metrics"
	"github.com/daffahilmyf/ride-hailing/services/location/internal/app/usecase"
	"github.com/daffahilmyf/ride-hailing/services/location/internal/infra"
	"github.com/google/uuid"
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
	Short: "Start location service",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := infra.LoadConfig()
		logger := infra.NewLogger()
		defer logger.Sync()

		if cfg.LocationKeyPrefix == "" || cfg.GeoKey == "" || cfg.RateLimitKeyPrefix == "" {
			logger.Fatal("config.invalid_location_keys")
		}

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
		logger.Info("redis.connected", zap.String("addr", cfg.RedisAddr))
		defer redisClient.Close()

		locMetrics := &metrics.LocationMetrics{}
		repo := redisadapter.NewLocationRepo(redisClient, cfg.LocationKeyPrefix, cfg.GeoKey, locMetrics)
		var limiter *redisadapter.RateLimiter
		if cfg.RateLimitEnabled {
			limiter = redisadapter.NewRateLimiter(redisClient)
		}

		var publisher *broker.Publisher
		var nc *nats.Conn
		if cfg.EventsEnabled {
			var err error
			nc, err = nats.Connect(cfg.NATSURL)
			if err != nil {
				logger.Fatal("nats.connect_failed", zap.Error(err))
			}
			logger.Info("nats.connected", zap.String("url", cfg.NATSURL))
			js, err := nc.JetStream()
			if err != nil {
				logger.Fatal("nats.jetstream_failed", zap.Error(err))
			}
			logger.Info("nats.jetstream_ready")
			ensureStream(logger, js, "DRIVERS", []string{"driver.>"}, cfg.NATSSelfHeal)
			publisher = broker.NewPublisher(js)
		}
		if nc != nil {
			defer nc.Close()
		}

		uc := &usecase.LocationService{
			Repo:           repo,
			Publisher:      publisher,
			RateLimiter:    limiter,
			PublishEnabled: cfg.EventsEnabled,
			LocationTTL:    time.Duration(cfg.LocationTTLSeconds) * time.Second,
			MinUpdateGap:   time.Duration(cfg.RateLimitMinGapMs) * time.Millisecond,
			RateKeyPrefix:  cfg.RateLimitKeyPrefix,
			Clock:          usecase.SystemClock{},
			IDGen:          uuid.NewString,
		}

		grpcMetrics := grpcadapter.NewMetrics()
		if cfg.Observability.MetricsEnabled {
			promMetrics := grpcadapter.NewPromMetrics(cfg.ServiceName)
			locPromMetrics := metrics.NewPromMetrics(cfg.ServiceName)
			registry := prometheus.NewRegistry()
			registry.MustRegister(promMetrics.Requests, promMetrics.Latency, locPromMetrics.StaleGeoRemoved)
			grpcMetrics.AttachProm(promMetrics)
			locMetrics.AttachProm(locPromMetrics)
			go serveMetrics(cfg.Observability.MetricsAddr, registry, logger)
		}

		if cfg.CleanupEnabled && cfg.CleanupIntervalSeconds > 0 {
			go func() {
				ticker := time.NewTicker(time.Duration(cfg.CleanupIntervalSeconds) * time.Second)
				defer ticker.Stop()
				cursor := uint64(0)
				for range ticker.C {
					next, removed, err := repo.CleanStaleGeo(context.Background(), cursor, cfg.CleanupBatchSize)
					if err != nil {
						logger.Warn("cleanup.stale_geo_failed", zap.Error(err))
					} else if removed > 0 {
						logger.Info("cleanup.stale_geo_removed", zap.Int("count", removed))
					}
					cursor = next
				}
			}()
		}
		srv := grpcadapter.NewServer(logger, handlers.Dependencies{Usecase: uc}, grpcMetrics, grpcadapter.AuthConfig{
			Enabled: cfg.InternalAuthEnabled,
			Token:   cfg.InternalAuthToken,
		})
		healthSrv := health.NewServer()
		healthpb.RegisterHealthServer(srv.GRPC(), healthSrv)
		reflection.Register(srv.GRPC())

		lis, err := net.Listen("tcp", cfg.GRPCAddr)
		if err != nil {
			logger.Fatal("grpc.listen_failed", zap.Error(err))
		}

		go func() {
			if err := srv.GRPC().Serve(lis); err != nil {
				logger.Fatal("grpc.serve_failed", zap.Error(err))
			}
		}()

		healthSrv.SetServingStatus("location.v1.LocationService", healthpb.HealthCheckResponse_SERVING)
		waitForShutdown(srv.GRPC(), cfg.ShutdownTimeoutSeconds, logger)
		return nil
	},
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

func ensureStream(logger *zap.Logger, js nats.JetStreamContext, name string, subjects []string, selfHeal bool) {
	if js == nil {
		return
	}
	info, err := js.StreamInfo(name)
	if err == nil {
		if !selfHeal {
			logger.Warn("nats.self_heal_disabled", zap.String("stream", name))
			return
		}
		existing := map[string]struct{}{}
		for _, s := range info.Config.Subjects {
			existing[s] = struct{}{}
		}
		updated := false
		for _, s := range subjects {
			if _, ok := existing[s]; !ok {
				info.Config.Subjects = append(info.Config.Subjects, s)
				updated = true
			}
		}
		if !updated {
			return
		}
		if _, err := js.UpdateStream(&info.Config); err != nil {
			logger.Warn("nats.stream_update_failed", zap.String("stream", name), zap.Error(err))
			return
		}
		logger.Info("nats.stream_updated", zap.String("stream", name))
		return
	}

	if !selfHeal {
		logger.Warn("nats.stream_missing", zap.String("stream", name))
		return
	}
	_, err = js.AddStream(&nats.StreamConfig{
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
