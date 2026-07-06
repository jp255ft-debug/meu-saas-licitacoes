package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// PoolConfig contém as configurações do pool de conexões PostgreSQL.
type PoolConfig struct {
	MaxConns    int32
	MinConns    int32
	MaxLifetime time.Duration
	MaxIdleTime time.Duration
}

// DefaultPoolConfig retorna a configuração padrão para produção.
func DefaultPoolConfig() PoolConfig {
	return PoolConfig{
		MaxConns:    20,
		MinConns:    5,
		MaxLifetime: 30 * time.Minute,
		MaxIdleTime: 5 * time.Minute,
	}
}

// TestPoolConfig retorna a configuração reduzida para testes.
func TestPoolConfig() PoolConfig {
	return PoolConfig{
		MaxConns:    5,
		MinConns:    1,
		MaxLifetime: 30 * time.Minute,
		MaxIdleTime: 5 * time.Minute,
	}
}

// NewPool cria um novo pool de conexões PostgreSQL com a configuração fornecida.
func NewPool(ctx context.Context, connString string, cfg PoolConfig) (*pgxpool.Pool, error) {
	config, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, fmt.Errorf("falha ao analisar string de conexão: %w", err)
	}

	config.MaxConns = cfg.MaxConns
	config.MinConns = cfg.MinConns
	config.MaxConnLifetime = cfg.MaxLifetime
	config.MaxConnIdleTime = cfg.MaxIdleTime

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("falha ao criar pool de conexões: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("falha ao conectar ao PostgreSQL: %w", err)
	}

	return pool, nil
}