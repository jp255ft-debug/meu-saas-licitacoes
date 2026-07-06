package outbound

import (
	"context"

	"github.com/jp255ft-debug/ledger-engine/internal/core/domain/entity"
)

// LedgerRepository define as operações de persistência do livro-razão.
type LedgerRepository interface {
	// CreateAccount persiste uma nova conta.
	CreateAccount(ctx context.Context, account *entity.Account) error

	// GetAccount busca uma conta pelo merchantID.
	GetAccount(ctx context.Context, merchantID string) (*entity.Account, error)

	// SaveEntry persiste uma entrada no livro-razão.
	SaveEntry(ctx context.Context, entry *entity.LedgerEntry) error

	// GetBalance calcula o saldo somando créditos e débitos.
	GetBalance(ctx context.Context, accountID string) (float64, error)

	// ExistsIdempotencyKey verifica se uma chave de idempotência já foi usada.
	ExistsIdempotencyKey(ctx context.Context, idempotencyKey string) (bool, error)
}