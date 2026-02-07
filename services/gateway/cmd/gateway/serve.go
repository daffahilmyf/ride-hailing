package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

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

		router := app.NewRouter(cfg, logger)
		server := &http.Server{
			Addr:    cfg.HTTPAddr,
			Handler: router,
		}

		go func() {
			if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				logger.Fatal("http.server.failed", zap.Error(err))
			}
		}()

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
