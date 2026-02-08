package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	dbadapter "github.com/daffahilmyf/ride-hailing/services/user/internal/adapters/db"
	grpcadapter "github.com/daffahilmyf/ride-hailing/services/user/internal/adapters/grpc"
	redisadapter "github.com/daffahilmyf/ride-hailing/services/user/internal/adapters/redis"
	"github.com/daffahilmyf/ride-hailing/services/user/internal/app/handlers"
	"github.com/daffahilmyf/ride-hailing/services/user/internal/app/metrics"
	"github.com/daffahilmyf/ride-hailing/services/user/internal/app/usecase"
	"github.com/daffahilmyf/ride-hailing/services/user/internal/infra"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start user service",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := infra.LoadConfig()
		logger := infra.NewLogger()
		defer logger.Sync()

		ctx := context.Background()
		store, err := dbadapter.NewPostgres(ctx, cfg.PostgresDSN)
		if err != nil {
			logger.Fatal("postgres.connect_failed", zap.Error(err))
		}

		repo := dbadapter.NewRepo(store.DB)
		uc := &usecase.Service{
			Repo:       repo,
			AuthConfig: cfg.Auth,
			SessionLimits: usecase.SessionLimitConfig{
				Rider:  cfg.SessionLimits.Rider,
				Driver: cfg.SessionLimits.Driver,
			},
		}

		redisClient := redis.NewClient(&redis.Options{
			Addr:     cfg.RedisAddr,
			Password: cfg.RedisPassword,
			DB:       cfg.RedisDB,
		})
		if err := redisClient.Ping(ctx).Err(); err != nil {
			logger.Warn("redis.connect_failed", zap.Error(err))
		}
		rLimiter := redisadapter.NewRateLimiter(redisClient)

		grpcSrv := grpcadapter.NewServer(logger, grpcadapter.Dependencies{Usecase: uc}, grpcadapter.AuthConfig{
			Enabled: cfg.InternalAuth.Enabled,
			Token:   cfg.InternalAuth.Token,
		})

		go func() {
			if err := grpcSrv.Serve(cfg.GRPCAddr); err != nil {
				logger.Fatal("grpc.serve_failed", zap.Error(err))
			}
		}()

		var authMetrics *metrics.AuthMetrics
		var registry *prometheus.Registry
		if cfg.Observability.MetricsEnabled {
			authMetrics = &metrics.AuthMetrics{}
			promMetrics := metrics.NewPromMetrics(cfg.ServiceName)
			registry = prometheus.NewRegistry()
			registry.MustRegister(promMetrics.Requests, promMetrics.Latency)
			authMetrics.AttachProm(promMetrics)
		}

		router := gin.New()
		router.Use(gin.Recovery())
		limiter := &handlers.RateLimiter{
			Redis:  rLimiter,
			Limit:  cfg.RateLimit.AuthRequests,
			Window: time.Duration(cfg.RateLimit.WindowSeconds) * time.Second,
			Prefix: cfg.RateLimit.KeyPrefix,
		}
		handlers.RegisterRoutes(router, uc, logger, authMetrics, limiter, cfg.InternalAuth.Enabled, cfg.InternalAuth.Token)

		httpSrv := &http.Server{
			Addr:         cfg.HTTPAddr,
			Handler:      router,
			ReadTimeout:  time.Duration(cfg.HTTPReadTimeoutSeconds) * time.Second,
			WriteTimeout: time.Duration(cfg.HTTPWriteTimeoutSeconds) * time.Second,
		}

		if cfg.Observability.MetricsEnabled && registry != nil {
			go serveMetrics(cfg.Observability.MetricsAddr, registry, logger)
		}

		go func() {
			if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				logger.Fatal("http.serve_failed", zap.Error(err))
			}
		}()

		waitForShutdown(httpSrv, grpcSrv.GRPC(), cfg.ShutdownTimeoutSeconds, logger)
		return nil
	},
}

func waitForShutdown(httpSrv *http.Server, grpcSrv *grpc.Server, timeoutSeconds int, logger *zap.Logger) {
	signals := []os.Signal{syscall.SIGINT, syscall.SIGTERM}
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, signals...)

	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeoutSeconds)*time.Second)
	defer cancel()

	httpDone := make(chan struct{})
	go func() {
		_ = httpSrv.Shutdown(ctx)
		close(httpDone)
	}()

	grpcDone := make(chan struct{})
	go func() {
		grpcSrv.GracefulStop()
		close(grpcDone)
	}()

	select {
	case <-httpDone:
	case <-ctx.Done():
		_ = httpSrv.Close()
	}

	select {
	case <-grpcDone:
	case <-ctx.Done():
		grpcSrv.Stop()
	}

	logger.Info("shutdown.complete")
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
