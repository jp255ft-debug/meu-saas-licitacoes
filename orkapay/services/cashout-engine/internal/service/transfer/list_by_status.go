package transfer

import (
	"context"
	"fmt"

	"github.com/jp255ft-debug/cashout-engine/internal/core/domain/entity"
	"github.com/jp255ft-debug/cashout-engine/internal/core/port/outbound"
)

// ListTransfersByStatusUseCase lista transferências por status.
type ListTransfersByStatusUseCase struct {
	repo outbound.TransferRepository
}

// NewListTransfersByStatusUseCase cria uma nova instância do caso de uso.
func NewListTransfersByStatusUseCase(repo outbound.TransferRepository) *ListTransfersByStatusUseCase {
	return &ListTransfersByStatusUseCase{repo: repo}
}

// Execute executa a listagem de transferências por status.
func (uc *ListTransfersByStatusUseCase) Execute(ctx context.Context, status string, limit, offset int) ([]*entity.Transfer, error) {
	transferStatus := entity.TransferStatus(status)
	if transferStatus == "" {
		return nil, fmt.Errorf("status inválido: %s", status)
	}

	transfers, err := uc.repo.ListByStatus(ctx, transferStatus, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("falha ao listar transferências: %w", err)
	}
	return transfers, nil
}