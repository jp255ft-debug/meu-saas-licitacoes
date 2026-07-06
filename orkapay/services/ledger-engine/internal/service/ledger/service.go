package ledger

import (
	"context"
	"fmt"

	"github.com/jp255ft-debug/ledger-engine/internal/core/domain"
	"github.com/jp255ft-debug/ledger-engine/internal/core/domain/entity"
	"github.com/jp255ft-debug/ledger-engine/internal/core/port/inbound"
	"github.com/jp255ft-debug/ledger-engine/internal/core/port/outbound"
)

// LedgerServiceImpl implementa inbound.LedgerService.
type LedgerServiceImpl struct {
	repo outbound.LedgerRepository
}

// NewLedgerService cria uma nova instância do serviço.
func NewLedgerService(repo outbound.LedgerRepository) *LedgerServiceImpl {
	return &LedgerServiceImpl{repo: repo}
}

func (s *LedgerServiceImpl) CreateAccount(ctx context.Context, merchantID, currency string) (*entity.Account, error) {
	if merchantID == "" || currency == "" {
		return nil, fmt.Errorf("merchantID e currency são obrigatórios")
	}

	account := entity.NewAccount(merchantID, currency)
	if err := s.repo.CreateAccount(ctx, account); err != nil {
		return nil, err
	}

	return account, nil
}

func (s *LedgerServiceImpl) GetAccount(ctx context.Context, merchantID string) (*entity.Account, error) {
	return s.repo.GetAccount(ctx, merchantID)
}

// Compile-time check
var _ inbound.LedgerService = (*LedgerServiceImpl)(nil)