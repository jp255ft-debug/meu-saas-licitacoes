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
	"github.com/jp255ft-debug/cashout-engine/internal/adapter/inbound/rest/middleware"
	"github.com/jp255ft-debug/cashout-engine/internal/adapter/inbound/rest/router"
	"github.com/jp255ft-debug/cashout-engine/internal/adapter/outbound/postgres"
	"github.com/jp255ft-debug/cashout-engine/internal/adapter/outbound/provider"
	"github.com/jp255ft-debug/cashout-engine/internal/service/transfer"
)

func main() {
	// 1. Carregar configuração
	cfg := LoadConfig()

	// 2. Configurar logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	logger.Info("iniciando cashout-engine", "port", cfg.ServerPort)

	// 3. Conectar ao banco de dados
	var svc *transfer.ServiceImpl

	if cfg.DatabaseURL != "" {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		pool, err := postgres.NewPool(ctx, cfg.DatabaseURL, postgres.DefaultPoolConfig())
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
		repo := postgres.NewTransferRepository(pool)
		prov := provider.NewStarkBankMock()

		svc = transfer.NewService(repo, prov)
		logger.Info("conectado ao banco de dados e repositório pronto")
	} else {
		logger.Warn("DATABASE_URL não configurada — modo sem persistência")
	}

	// 6. Configurar rate limiter e router HTTP
	apiKeys := parseAPIKeys(cfg.APIKeys)
	rateLimiter := middleware.NewRateLimiter(100, 1*time.Minute)
	httpHandler := router.New(svc, apiKeys, rateLimiter)

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
	CREATE TABLE IF NOT EXISTS transfers (
		id VARCHAR(64) PRIMARY KEY,
		external_id VARCHAR(128),
		status VARCHAR(32) NOT NULL DEFAULT 'PENDING',
		amount_cents BIGINT NOT NULL,
		payee_name VARCHAR(255) NOT NULL,
		payee_tax_id VARCHAR(32) NOT NULL,
		payee_pix_key VARCHAR(128) NOT NULL,
		payee_pix_type VARCHAR(16),
		payee_bank_code VARCHAR(16),
		payee_branch VARCHAR(16),
		payee_account VARCHAR(32),
		payee_account_type VARCHAR(16),
		provider VARCHAR(32) NOT NULL,
		description TEXT NOT NULL DEFAULT '',
		metadata JSONB DEFAULT '{}',
		processed_at TIMESTAMPTZ,
		settled_at TIMESTAMPTZ,
		failed_at TIMESTAMPTZ,
		cancelled_at TIMESTAMPTZ,
		failure_reason TEXT,
		version INT NOT NULL DEFAULT 1,
		created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
	);
	CREATE INDEX IF NOT EXISTS idx_transfers_status ON transfers(status);
	CREATE INDEX IF NOT EXISTS idx_transfers_external_id ON transfers(external_id);
	CREATE INDEX IF NOT EXISTS idx_transfers_payee_tax_id ON transfers(payee_tax_id);
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