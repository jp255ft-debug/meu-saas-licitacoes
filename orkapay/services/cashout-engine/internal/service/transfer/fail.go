package transfer

import (
	"context"
	"fmt"

	"github.com/jp255ft-debug/cashout-engine/internal/core/port/inbound"
	"github.com/jp255ft-debug/cashout-engine/internal/core/port/outbound"
)

// FailTransferUseCase orquestra a marcação de falha de uma transferência.
type FailTransferUseCase struct {
	repo outbound.TransferRepository
}

// NewFailTransferUseCase cria uma nova instância do caso de uso.
func NewFailTransferUseCase(repo outbound.TransferRepository) *FailTransferUseCase {
	return &FailTransferUseCase{repo: repo}
}

// Execute executa o fluxo de falha de transferência.
func (uc *FailTransferUseCase) Execute(ctx context.Context, id, reason string) (*inbound.TransferOutput, error) {
	// 1. Buscar transferência
	transfer, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("transferência não encontrada: %w", err)
	}

	// 2. Marcar como falha no domínio
	if err := transfer.Fail(reason); err != nil {
		return nil, fmt.Errorf("falha ao marcar transferência como falha: %w", err)
	}

	// 3. Persistir com optimistic locking
	if err := uc.repo.Update(ctx, transfer); err != nil {
		return nil, fmt.Errorf("falha ao atualizar transferência: %w", err)
	}

	return MapToTransferOutput(transfer), nil
}