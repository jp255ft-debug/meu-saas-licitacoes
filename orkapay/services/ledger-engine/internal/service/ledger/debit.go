package ledger

import (
	"context"
	"fmt"

	"github.com/jp255ft-debug/ledger-engine/internal/core/domain"
	"github.com/jp255ft-debug/ledger-engine/internal/core/domain/entity"
)

// Debit debita um valor da conta do merchant, validando saldo.
func (s *LedgerServiceImpl) Debit(ctx context.Context, merchantID, transactionID, operationType, idempotencyKey string, amount float64) (*entity.LedgerEntry, error) {
	if amount <= 0 {
		return nil, domain.ErrInvalidAmount
	}

	// Verificar idempotência
	exists, err := s.repo.ExistsIdempotencyKey(ctx, idempotencyKey)
	if err != nil {
		return nil, fmt.Errorf("falha ao verificar idempotência: %w", err)
	}
	if exists {
		return nil, domain.ErrDuplicateIdempotency
	}

	account, err := s.repo.GetAccount(ctx, merchantID)
	if err != nil {
		return nil, err
	}

	if account.Status != entity.AccountStatusActive {
		return nil, domain.ErrAccountBlocked
	}

	// Validar saldo antes de debitar
	balance, err := s.repo.GetBalance(ctx, account.ID)
	if err != nil {
		return nil, fmt.Errorf("falha ao consultar saldo: %w", err)
	}
	if balance < amount {
		return nil, domain.ErrInsufficientBalance
	}

	entry := entity.NewDebitEntry(account.ID, transactionID, operationType, idempotencyKey, amount)
	if err := s.repo.SaveEntry(ctx, entry); err != nil {
		return nil, fmt.Errorf("falha ao debitar: %w", err)
	}

	return entry, nil
}