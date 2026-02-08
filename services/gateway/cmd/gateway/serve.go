package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/daffahilmyf/ride-hailing/services/gateway/internal/adapters/cache"
	grpcadapter "github.com/daffahilmyf/ride-hailing/services/gateway/internal/adapters/grpc"
	"github.com/daffahilmyf/ride-hailing/services/gateway/internal/app"
	"github.com/daffahilmyf/ride-hailing/services/gateway/internal/infra"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the API gateway",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := infra.LoadConfig()
		logger := infra.NewLogger()
		defer logger.Sync()

		if err := infra.ValidateConfig(cfg); err != nil {
			logger.Fatal("config.invalid", zap.Error(err))
		}

		shutdownTelemetry, err := infra.SetupTelemetry(context.Background(), cfg)
		if err != nil {
			logger.Fatal("telemetry.init_failed", zap.Error(err))
		}
		defer shutdownTelemetry(context.Background())

		var grpcClients *grpcadapter.Clients
		if cfg.GRPC.ConnectRequired {
			clients, err := grpcadapter.NewClients(context.Background(), cfg.GRPC)
			if err != nil {
				logger.Fatal("grpc.clients.failed", zap.Error(err))
			}
			logger.Info("grpc.clients.connected")
			grpcClients = clients
		} else {
			clients, err := grpcadapter.NewClients(context.Background(), cfg.GRPC)
			if err != nil {
				logger.Warn("grpc.clients.failed", zap.Error(err))
			}
			logger.Info("grpc.clients.ready")
			grpcClients = clients
		}
		if grpcClients != nil {
			defer grpcClients.Close()
		}

		redisClient := cache.NewRedisClient(cache.RedisConfig{
			Addr:     cfg.Redis.Addr,
			Password: cfg.Redis.Password,
			DB:       cfg.Redis.DB,
		})
		if err := redisClient.Ping(context.Background()).Err(); err != nil {
			logger.Fatal("redis.connect_failed", zap.Error(err))
		}
		logger.Info("redis.connected", zap.String("addr", cfg.Redis.Addr))
		defer redisClient.Close()

		deps := app.Deps{
			RideClient:     grpcClients.RideClient,
			MatchingClient: grpcClients.MatchingClient,
			LocationClient: grpcClients.LocationClient,
			AuthClient:     grpcClients.AuthClient,
		}

		router := app.NewRouter(cfg, logger, deps, redisClient, grpcClients)
		server := &http.Server{
			Addr:    cfg.HTTPAddr,
			Handler: router,
		}

		go func() {
			if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				logger.Fatal("http.server.failed", zap.Error(err))
			}
		}()
		logger.Info("service.started", zap.String("service", cfg.ServiceName), zap.String("http_addr", cfg.HTTPAddr))

		waitForShutdown(server, cfg.ShutdownTimeoutSeconds, logger)
		return nil
	},
}

func waitForShutdown(server *http.Server, timeoutSeconds int, logger *zap.Logger) {
	signals := []os.Signal{syscall.SIGINT, syscall.SIGTERM}
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, signals...)

	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeoutSeconds)*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Error("http.server.shutdown_failed", zap.Error(err))
		return
	}
	logger.Info("http.server.shutdown_complete")
}
