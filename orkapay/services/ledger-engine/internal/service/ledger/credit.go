package ledger

import (
	"context"
	"fmt"

	"github.com/jp255ft-debug/ledger-engine/internal/core/domain"
	"github.com/jp255ft-debug/ledger-engine/internal/core/domain/entity"
)

// Credit credita um valor na conta do merchant.
// Cria a conta automaticamente se não existir.
func (s *LedgerServiceImpl) Credit(ctx context.Context, merchantID, transactionID, operationType, idempotencyKey string, amount float64) (*entity.LedgerEntry, error) {
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

	// Buscar ou criar conta
	account, err := s.repo.GetAccount(ctx, merchantID)
	if err != nil {
		if err == domain.ErrAccountNotFound {
			account = entity.NewAccount(merchantID, "BRL")
			if err := s.repo.CreateAccount(ctx, account); err != nil {
				return nil, fmt.Errorf("falha ao criar conta automaticamente: %w", err)
			}
		} else {
			return nil, err
		}
	}

	if account.Status != entity.AccountStatusActive {
		return nil, domain.ErrAccountBlocked
	}

	entry := entity.NewCreditEntry(account.ID, transactionID, operationType, idempotencyKey, amount)
	if err := s.repo.SaveEntry(ctx, entry); err != nil {
		return nil, fmt.Errorf("falha ao creditar: %w", err)
	}

	return entry, nil
}