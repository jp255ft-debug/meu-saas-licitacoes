package ledger

import (
	"context"
	"fmt"
)

// GetBalance retorna o saldo atual de um merchant.
func (s *LedgerServiceImpl) GetBalance(ctx context.Context, merchantID string) (float64, error) {
	account, err := s.repo.GetAccount(ctx, merchantID)
	if err != nil {
		return 0, err
	}

	balance, err := s.repo.GetBalance(ctx, account.ID)
	if err != nil {
		return 0, fmt.Errorf("falha ao calcular saldo: %w", err)
	}

	return balance, nil
}