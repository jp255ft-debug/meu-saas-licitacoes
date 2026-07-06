package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jp255ft-debug/ledger-engine/internal/core/domain"
	"github.com/jp255ft-debug/ledger-engine/internal/core/domain/entity"
	"github.com/jp255ft-debug/ledger-engine/internal/core/port/outbound"
)

// LedgerRepository implementa outbound.LedgerRepository usando PostgreSQL.
type LedgerRepository struct {
	pool *pgxpool.Pool
}

// NewLedgerRepository cria uma nova instância do repositório.
func NewLedgerRepository(pool *pgxpool.Pool) *LedgerRepository {
	return &LedgerRepository{pool: pool}
}

func (r *LedgerRepository) CreateAccount(ctx context.Context, account *entity.Account) error {
	query := `
		INSERT INTO ledger_accounts (id, merchant_id, currency, status)
		VALUES ($1, $2, $3, $4)
	`
	_, err := r.pool.Exec(ctx, query, account.ID, account.MerchantID, account.Currency, string(account.Status))
	if err != nil {
		return fmt.Errorf("falha ao criar conta: %w", err)
	}
	return nil
}

func (r *LedgerRepository) GetAccount(ctx context.Context, merchantID string) (*entity.Account, error) {
	query := `SELECT id, merchant_id, currency, status, created_at, updated_at FROM ledger_accounts WHERE merchant_id = $1`
	row := r.pool.QueryRow(ctx, query, merchantID)

	var a accountRow
	if err := row.Scan(&a.ID, &a.MerchantID, &a.Currency, &a.Status, &a.CreatedAt, &a.UpdatedAt); err != nil {
		return nil, domain.ErrAccountNotFound
	}

	return a.toDomain(), nil
}

func (r *LedgerRepository) SaveEntry(ctx context.Context, entry *entity.LedgerEntry) error {
	query := `
		INSERT INTO ledger_entries (id, account_id, transaction_id, entry_type, amount, operation_type, idempotency_key)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := r.pool.Exec(ctx, query,
		entry.ID, entry.AccountID, entry.TransactionID,
		string(entry.EntryType), entry.Amount, entry.OperationType, entry.IdempotencyKey,
	)
	if err != nil {
		return fmt.Errorf("falha ao salvar entrada ledger: %w", err)
	}
	return nil
}

func (r *LedgerRepository) GetBalance(ctx context.Context, accountID string) (float64, error) {
	query := `
		SELECT COALESCE(SUM(CASE WHEN entry_type = 'CREDIT' THEN amount ELSE -amount END), 0)
		FROM ledger_entries
		WHERE account_id = $1
	`
	var balance float64
	if err := r.pool.QueryRow(ctx, query, accountID).Scan(&balance); err != nil {
		return 0, fmt.Errorf("falha ao calcular saldo: %w", err)
	}
	return balance, nil
}

func (r *LedgerRepository) ExistsIdempotencyKey(ctx context.Context, idempotencyKey string) (bool, error) {
	query := `SELECT COUNT(*) FROM ledger_entries WHERE idempotency_key = $1`
	var count int64
	if err := r.pool.QueryRow(ctx, query, idempotencyKey).Scan(&count); err != nil {
		return false, fmt.Errorf("falha ao verificar idempotência: %w", err)
	}
	return count > 0, nil
}

// Compile-time check
var _ outbound.LedgerRepository = (*LedgerRepository)(nil)