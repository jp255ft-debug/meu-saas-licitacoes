package transfer

import (
	"context"
	"fmt"

	"github.com/jp255ft-debug/cashout-engine/internal/core/domain/entity"
	"github.com/jp255ft-debug/cashout-engine/internal/core/port/inbound"
	"github.com/jp255ft-debug/cashout-engine/internal/core/port/outbound"
)

// SettleTransferUseCase orquestra a liquidação de uma transferência.
type SettleTransferUseCase struct {
	repo outbound.TransferRepository
}

// NewSettleTransferUseCase cria uma nova instância do caso de uso.
func NewSettleTransferUseCase(repo outbound.TransferRepository) *SettleTransferUseCase {
	return &SettleTransferUseCase{repo: repo}
}

// Execute executa o fluxo de liquidação de transferência.
func (uc *SettleTransferUseCase) Execute(ctx context.Context, id string) (*inbound.TransferOutput, error) {
	// 1. Buscar transferência
	transfer, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("transferência não encontrada: %w", err)
	}

	// 2. Verificar se está aprovada
	if transfer.Status() != entity.TransferStatusApproved {
		return nil, fmt.Errorf("transferência não está aprovada: status=%s", transfer.Status())
	}

	// 3. Liquidar no domínio
	if err := transfer.Settle(); err != nil {
		return nil, fmt.Errorf("falha ao liquidar transferência: %w", err)
	}

	// 4. Persistir com optimistic locking
	if err := uc.repo.Update(ctx, transfer); err != nil {
		return nil, fmt.Errorf("falha ao atualizar transferência: %w", err)
	}

	return MapToTransferOutput(transfer), nil
}