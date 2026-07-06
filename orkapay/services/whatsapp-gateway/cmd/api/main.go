package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jp255ft-debug/whatsapp-gateway/internal/adapter/inbound/rest/handler"
	"github.com/jp255ft-debug/whatsapp-gateway/internal/adapter/outbound/bsp"
	"github.com/jp255ft-debug/whatsapp-gateway/internal/config"
	"github.com/jp255ft-debug/whatsapp-gateway/internal/core/service/notification"
	"github.com/jp255ft-debug/whatsapp-gateway/internal/core/service/whatsapp"
)

func main() {
	// 1. Load config
	cfg := config.Load()
	_ = cfg // usado futuramente para BSP real

	// 2. Setup logger
	opts := &slog.HandlerOptions{Level: slog.LevelInfo}
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, opts)))

	// 3. Initialize dependencies
	bspSender := bsp.NewMockSender()

	whatsappService := whatsapp.NewService(bspSender)

	// 4. Setup HTTP router
	mux := http.NewServeMux()

	// Health checks
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})
	mux.HandleFunc("GET /ready", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ready"}`))
	})

	// Webhook route (receives messages from WhatsApp BSP)
	webhookHandler := handler.NewWebhookHandler(whatsappService)
	mux.HandleFunc("POST /webhook/whatsapp", webhookHandler.HandleWebhook)

	// Notification route (receives payment confirmation from reconciliation)
	notificationService := notification.NewService(bspSender)
	notificationHandler := handler.NewNotificationHandler(notificationService)
	mux.HandleFunc("POST /api/v1/notifications/payment", notificationHandler.SendPaymentConfirmation)

	// 5. Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8082"
	}

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  30 * time.Second,
	}

	// Graceful shutdown
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGTERM)

	go func() {
		slog.Info("starting whatsapp-gateway", "port", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server failed", "error", err)
			os.Exit(1)
		}
	}()

	<-done
	slog.Info("shutting down gracefully...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("shutdown failed", "error", err)
	}
	slog.Info("shutdown complete")
}