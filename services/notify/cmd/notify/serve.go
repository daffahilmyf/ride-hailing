package main

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/daffahilmyf/ride-hailing/services/notify/internal/adapters/broker"
	"github.com/daffahilmyf/ride-hailing/services/notify/internal/app"
	"github.com/daffahilmyf/ride-hailing/services/notify/internal/app/workers"
	"github.com/daffahilmyf/ride-hailing/services/notify/internal/infra"
	"github.com/nats-io/nats.go"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start notification service",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := infra.LoadConfig()
		logger := infra.NewLogger()
		defer logger.Sync()

		hub := app.NewHub(cfg.SSEBufferSize)

		mux := http.NewServeMux()
		mux.HandleFunc("/sse", func(w http.ResponseWriter, r *http.Request) {
			userID := r.Header.Get("X-User-Id")
			role := r.Header.Get("X-Role")
			if userID == "" || role == "" {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			w.Header().Set("Content-Type", "text/event-stream")
			w.Header().Set("Cache-Control", "no-cache")
			w.Header().Set("Connection", "keep-alive")

			flusher, ok := w.(http.Flusher)
			if !ok {
				http.Error(w, "streaming unsupported", http.StatusInternalServerError)
				return
			}

			id := r.RemoteAddr + "-" + time.Now().Format(time.RFC3339Nano)
			filters := app.Filters{RideID: r.URL.Query().Get("ride_id")}
			switch role {
			case "driver":
				filters.DriverID = userID
			case "rider":
				filters.RiderID = userID
			default:
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			sub := hub.Subscribe(id, filters)
			defer hub.Unsubscribe(id)

			ctx := r.Context()
			keepAlive := time.Duration(cfg.SSEKeepaliveSeconds) * time.Second
			go app.KeepAlive(w, keepAlive)

			for {
				select {
				case <-ctx.Done():
					return
				case n, ok := <-sub.Ch:
					if !ok {
						return
					}
					if err := app.WriteSSE(w, n); err != nil {
						return
					}
					flusher.Flush()
				}
			}
		})

		server := &http.Server{
			Addr:              cfg.HTTPAddr,
			Handler:           mux,
			ReadHeaderTimeout: 3 * time.Second,
		}

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

			ensureStream(logger, js, "RIDES", []string{"ride.>"}, cfg.NATSSelfHeal)
			ensureStream(logger, js, "DRIVERS", []string{"driver.>"}, cfg.NATSSelfHeal)

			consumer := broker.NewConsumer(js)
			handler := func(subject string, payload []byte) error {
				n, data, err := toNotification(subject, payload)
				if err != nil {
					return err
				}
				hub.Broadcast(n, data)
				return nil
			}

			rideConsumer := &workers.EventConsumer{
				Consumer: consumer,
				Subject:  cfg.RideSubject,
				Durable:  "notify-ride-events",
				Batch:    20,
				Logger:   logger,
				Handler: func(ctx context.Context, payload []byte) error {
					return handler(cfg.RideSubject, payload)
				},
			}
			go func() {
				if err := rideConsumer.Run(ctx); err != nil {
					logger.Warn("event.consumer_stopped", zap.String("subject", cfg.RideSubject), zap.Error(err))
				}
			}()

			driverConsumer := &workers.EventConsumer{
				Consumer: consumer,
				Subject:  cfg.DriverSubject,
				Durable:  "notify-driver-events",
				Batch:    20,
				Logger:   logger,
				Handler: func(ctx context.Context, payload []byte) error {
					return handler(cfg.DriverSubject, payload)
				},
			}
			go func() {
				if err := driverConsumer.Run(ctx); err != nil {
					logger.Warn("event.consumer_stopped", zap.String("subject", cfg.DriverSubject), zap.Error(err))
				}
			}()
		}

		go func() {
			logger.Info("http.started", zap.String("addr", cfg.HTTPAddr))
			if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
				logger.Fatal("http.listen_failed", zap.Error(err))
			}
		}()

		stop := make(chan os.Signal, 1)
		signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
		<-stop
		logger.Info("shutdown.signal")

		ctxTimeout, cancelTimeout := context.WithTimeout(context.Background(), time.Duration(cfg.ShutdownTimeoutSeconds)*time.Second)
		defer cancelTimeout()
		_ = server.Shutdown(ctxTimeout)
		logger.Info("shutdown.complete")
		return nil
	},
}

func toNotification(subject string, payload []byte) (app.Notification, map[string]interface{}, error) {
	var envelope map[string]interface{}
	if err := json.Unmarshal(payload, &envelope); err != nil {
		return app.Notification{}, nil, err
	}

	data, _ := envelope["data"].(map[string]interface{})
	if data == nil {
		data, _ = envelope["payload"].(map[string]interface{})
	}
	if data == nil {
		data = map[string]interface{}{}
	}

	eventType, _ := envelope["type"].(string)
	traceID, _ := envelope["trace_id"].(string)
	requestID, _ := envelope["request_id"].(string)
	if eventType == "" {
		eventType = subject
	}

	n := app.Notification{
		Event:     eventType,
		Subject:   subject,
		TraceID:   traceID,
		RequestID: requestID,
		Data:      data,
		Timestamp: time.Now().UTC().Format(time.RFC3339Nano),
	}
	return n, data, nil
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
