package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jp255ft-debug/payment-core/internal/adapter/inbound/rest/handler"
	"github.com/jp255ft-debug/payment-core/internal/adapter/inbound/rest/router"
	"github.com/jp255ft-debug/payment-core/internal/adapter/outbound/messaging/kafka"
	"github.com/jp255ft-debug/payment-core/internal/adapter/outbound/postgres"
	"github.com/jp255ft-debug/payment-core/internal/core/application"
)

func main() {
	ctx := context.Background()

	// ConfiguraÃ§Ãµes (hardcoded para MVP â€” usar env vars em produÃ§Ã£o)
	dbCfg := postgres.DBConfig{
		Host:            getEnv("DB_HOST", "localhost"),
		Port:            getEnv("DB_PORT", "5432"),
		User:            getEnv("DB_USER", "postgres"),
		Password:        getEnv("DB_PASSWORD", "postgres"),
		Database:        getEnv("DB_NAME", "orkapay"),
		SSLMode:         getEnv("DB_SSLMODE", "disable"),
		MaxConns:        50,
		MinConns:        10,
		MaxConnLifetime: 30 * time.Minute,
		MaxConnIdleTime: 5 * time.Minute,
	}

	// Conectar ao banco
	pool, err := postgres.NewDB(dbCfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Falha ao conectar ao banco: %v\n", err)
		os.Exit(1)
	}
	defer pool.Close()
	fmt.Println("âœ… Conectado ao PostgreSQL")

	// RepositÃ³rio
	repo := postgres.NewPaymentRepository(pool)

	// Mock de Kafka (MVP)
	pub := kafka.NewMockEventPublisher()

	// --- Gateway de Pagamento ---
	// Se ASAAS_API_KEY estiver configurada, usa cliente real; senao, usa mock
	var gw port.PaymentGateway
	asaasAPIKey := getEnv("ASAAS_API_KEY", "")
	if asaasAPIKey != "" {
		asaasBaseURL := getEnv("ASAAS_BASE_URL", "https://sandbox.asaas.com/api/v3")
		asaasClient := asaas.NewClient(asaasBaseURL, asaasAPIKey, 30*time.Second)
		gw = asaas.NewPaymentGateway(asaasClient)
		fmt.Println("✅ Gateway Asaas real configurado")
	} else {
		gw = port.NewMockPaymentGateway()
		fmt.Println("ℹ️ Usando gateway mock (ASAAS_API_KEY nao configurada)")
	}

	// Casos de uso
	createUC := application.NewCreatePaymentUseCase(repo, pub)
	processorUC := application.NewProcessPaymentUseCase(repo, pub, gw)
	cancelUC := application.NewCancelPaymentUseCase(repo, pub, gw)
	getUC := application.NewGetPaymentUseCase(repo)
	webhookUC := application.NewWebhookPaymentUseCase(repo, pub)

	// Handlers
	paymentHandler := handler.NewPaymentHandler(createUC, processorUC, cancelUC, getUC, webhookUC)
	healthHandler := handler.NewHealthHandler()

	// Router
	r := router.New(paymentHandler, healthHandler)

	// Server
	port := getEnv("APP_PORT", "8080")
	server := &http.Server{
		Addr:         ":" + port,
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan

		shutdownCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()

		fmt.Println("\nðŸ›‘ Encerrando servidor...")
		if err := server.Shutdown(shutdownCtx); err != nil {
			fmt.Fprintf(os.Stderr, "Erro no shutdown: %v\n", err)
		}
	}()

	fmt.Printf("ðŸš€ Servidor iniciado na porta %s\n", port)
	fmt.Println("Endpoints:")
	fmt.Println("  POST /api/v1/payments       - Criar pagamento")
	fmt.Println("  GET  /api/v1/payments/{id}  - Buscar pagamento")
	fmt.Println("  POST /api/v1/payments/{id}/cancel - Cancelar pagamento")
	fmt.Println("  POST /api/v1/webhooks/asaas - Webhook Asaas")
	fmt.Println("  GET  /health                - Health check")
	fmt.Println("  GET  /ready                 - Readiness check")

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		fmt.Fprintf(os.Stderr, "Falha ao iniciar servidor: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("ðŸ‘‹ Servidor encerrado")
}

// getEnv retorna o valor da variÃ¡vel de ambiente ou o default.
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
