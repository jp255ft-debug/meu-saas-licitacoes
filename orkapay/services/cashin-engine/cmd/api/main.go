package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jp255ft-debug/cashin-engine/internal/adapter/inbound/rest/handler"
	"github.com/jp255ft-debug/cashin-engine/internal/adapter/inbound/rest/middleware"
	"github.com/jp255ft-debug/cashin-engine/internal/adapter/inbound/rest/router"
	"github.com/jp255ft-debug/cashin-engine/internal/adapter/outbound/postgres"
	"github.com/jp255ft-debug/cashin-engine/internal/adapter/outbound/provider"
	"github.com/jp255ft-debug/cashin-engine/internal/core/port/outbound"
	"github.com/jp255ft-debug/cashin-engine/internal/service/invoice"
	"github.com/jp255ft-debug/cashin-engine/internal/service/webhook"
)

func main() {
	// 1. Carregar configuração
	cfg := LoadConfig()

	// 2. Configurar logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	logger.Info("iniciando cashin-engine", "port", cfg.ServerPort)

	// 3. Conectar ao banco de dados
	var svc *invoice.ServiceImpl
	var webhookSvc *webhook.WebhookService
	var pool *pgxpool.Pool

	if cfg.DatabaseURL != "" {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		var err error
		pool, err = postgres.NewPool(ctx, cfg.DatabaseURL, postgres.DefaultPoolConfig())
		if err != nil {
			logger.Error("falha ao conectar ao banco de dados", "error", err)
			os.Exit(1)
		}
		defer pool.Close()

		// 4. Executar migrations inline
		if err := runMigrations(ctx, pool); err != nil {
			logger.Error("falha ao executar migrations", "error", err)
			os.Exit(1)
		}

		// 5. Montar dependências
		repo := postgres.NewInvoiceRepository(pool)
		webhookRepo := postgres.NewWebhookRepository(pool)

		// Usar gateway real se configurado, senão usar mock
		asaasBaseURL := os.Getenv("ASAAS_BASE_URL")
		asaasAPIKey := os.Getenv("ASAAS_API_KEY")
		asaasWalletID := os.Getenv("ASAAS_WALLET_ID")

		var prov outbound.CashInProviderGateway
		if asaasAPIKey != "" && asaasBaseURL != "" {
			prov = provider.NewAsaasGateway(asaasBaseURL, asaasAPIKey, asaasWalletID)
			logger.Info("usando gateway Asaas real com Split de Pagamento")
		} else {
			prov = provider.NewAsaasMock()
			logger.Warn("ASAAS_API_KEY não configurada — usando mock Asaas")
		}

		webhookSvc = webhook.NewWebhookService(webhookRepo)
		svc = invoice.NewService(repo, prov, webhookSvc)
		logger.Info("conectado ao banco de dados e repositório pronto")
	} else {
		logger.Warn("DATABASE_URL não configurada — modo sem persistência")
	}

	// 6. Configurar autenticação, rate limiter e router HTTP
	apiKeys := parseAPIKeys(cfg.APIKeys)
	mapProvider := middleware.NewMapAuthProvider(apiKeys)
	
	var authProviders []middleware.AuthProvider
	authProviders = append(authProviders, mapProvider)
	
	// Se tiver pool de banco, adiciona CachedDBProvider como fallback
	if pool != nil {
		dbProvider := middleware.NewCachedDBProvider(func(apiKey string) (string, bool) {
			var clientID string
			err := pool.QueryRow(context.Background(),
				`SELECT client_id FROM api_keys WHERE key_hash = crypt($1, key_hash)`, apiKey,
			).Scan(&clientID)
			if err != nil {
				return "", false
			}
			return clientID, true
		})
		authProviders = append(authProviders, dbProvider)
	}

	// 6a. Instanciar Asaas webhook handler com validação HMAC
	asaasWebhookHandler := handler.NewAsaasWebhookHandler(cfg.AsaasWebhookToken, cfg.AsaasWebhookSecret)

	rateLimiter := middleware.NewRateLimiter(100, 1*time.Minute)
	httpHandler := router.New(svc, webhookSvc, authProviders, rateLimiter, asaasWebhookHandler)

	// 7. Servidor HTTP
	server := &http.Server{
		Addr:         ":" + cfg.ServerPort,
		Handler:      httpHandler,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// 8. Graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan

		logger.Info("recebido sinal de desligamento")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			logger.Error("erro no desligamento do servidor", "error", err)
		}
	}()

	logger.Info("servidor ouvindo", "addr", server.Addr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Error("erro no servidor", "error", err)
		os.Exit(1)
	}

	logger.Info("servidor desligado com sucesso")
}

// runMigrations executa o SQL de criação da tabela.
func runMigrations(ctx context.Context, pool *pgxpool.Pool) error {
	migration := `
	CREATE TABLE IF NOT EXISTS invoices (
		id VARCHAR(64) PRIMARY KEY,
		external_id VARCHAR(128),
		invoice_number VARCHAR(128) NOT NULL,
		status VARCHAR(32) NOT NULL DEFAULT 'PENDING',
		amount_cents BIGINT NOT NULL,
		due_date TIMESTAMPTZ NOT NULL,
		payer_name VARCHAR(255) NOT NULL,
		payer_tax_id VARCHAR(32) NOT NULL,
		payer_email VARCHAR(255),
		payer_phone VARCHAR(32),
		payer_address_line1 VARCHAR(255),
		payer_address_line2 VARCHAR(255),
		payer_city VARCHAR(128),
		payer_state VARCHAR(32),
		payer_zip_code VARCHAR(16),
		provider VARCHAR(32) NOT NULL,
		pix_qr_code TEXT,
		pix_key VARCHAR(128),
		boleto_barcode VARCHAR(64),
		boleto_digitable VARCHAR(64),
		metadata JSONB DEFAULT '{}',
		paid_amount_cents BIGINT NOT NULL DEFAULT 0,
		paid_at TIMESTAMPTZ,
		payment_method VARCHAR(32),
		cancelled_at TIMESTAMPTZ,
		failure_reason TEXT,
		version INT NOT NULL DEFAULT 1,
		created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
	);
	CREATE INDEX IF NOT EXISTS idx_invoices_status ON invoices(status);
	CREATE INDEX IF NOT EXISTS idx_invoices_external_id ON invoices(external_id);
	CREATE INDEX IF NOT EXISTS idx_invoices_payer_tax_id ON invoices(payer_tax_id);
	`
	_, err := pool.Exec(ctx, migration)
	return err
}

// parseAPIKeys converte string "key1:id1,key2:id2" em map.
func parseAPIKeys(raw string) map[string]string {
	result := make(map[string]string)
	if raw == "" {
		return result
	}
	for _, pair := range strings.Split(raw, ",") {
		parts := strings.SplitN(pair, ":", 2)
		if len(parts) == 2 {
			result[parts[0]] = parts[1]
		}
	}
	return result
}