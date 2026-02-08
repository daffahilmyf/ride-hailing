package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/daffahilmyf/ride-hailing/services/user/internal/adapters/db"
	grpcadapter "github.com/daffahilmyf/ride-hailing/services/user/internal/adapters/grpc"
	"github.com/daffahilmyf/ride-hailing/services/user/internal/app/handlers"
	"github.com/daffahilmyf/ride-hailing/services/user/internal/app/usecase"
	"github.com/daffahilmyf/ride-hailing/services/user/internal/infra"
	"github.com/gin-gonic/gin"
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
		store, err := db.NewPostgres(ctx, cfg.PostgresDSN)
		if err != nil {
			logger.Fatal("postgres.connect_failed", zap.Error(err))
		}

		repo := db.NewRepo(store.DB)
		uc := &usecase.Service{
			Repo:       repo,
			AuthConfig: cfg.Auth,
		}

		grpcSrv := grpcadapter.NewServer(logger, grpcadapter.Dependencies{Usecase: uc}, grpcadapter.AuthConfig{
			Enabled: cfg.InternalAuth.Enabled,
			Token:   cfg.InternalAuth.Token,
		})

		go func() {
			if err := grpcSrv.Serve(cfg.GRPCAddr); err != nil {
				logger.Fatal("grpc.serve_failed", zap.Error(err))
			}
		}()

		router := gin.New()
		router.Use(gin.Recovery())
		handlers.RegisterRoutes(router, uc)
		httpSrv := &http.Server{
			Addr:         cfg.HTTPAddr,
			Handler:      router,
			ReadTimeout:  time.Duration(cfg.HTTPReadTimeoutSeconds) * time.Second,
			WriteTimeout: time.Duration(cfg.HTTPWriteTimeoutSeconds) * time.Second,
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
