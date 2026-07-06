package inbound

import (
	"context"

	"github.com/jp255ft-debug/ledger-engine/internal/core/domain/entity"
)

// LedgerService define as operações do livro-razão.
type LedgerService interface {
	// Credit credita um valor na conta do merchant.
	Credit(ctx context.Context, merchantID, transactionID, operationType, idempotencyKey string, amount float64) (*entity.LedgerEntry, error)

	// Debit debita um valor da conta do merchant, validando saldo.
	Debit(ctx context.Context, merchantID, transactionID, operationType, idempotencyKey string, amount float64) (*entity.LedgerEntry, error)

	// GetBalance retorna o saldo atual de um merchant.
	GetBalance(ctx context.Context, merchantID string) (float64, error)

	// CreateAccount cria uma nova conta para um merchant.
	CreateAccount(ctx context.Context, merchantID, currency string) (*entity.Account, error)

	// GetAccount retorna a conta de um merchant.
	GetAccount(ctx context.Context, merchantID string) (*entity.Account, error)
}