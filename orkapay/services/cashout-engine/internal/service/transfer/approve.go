package transfer

import (
	"context"
	"fmt"

	"github.com/jp255ft-debug/cashout-engine/internal/core/domain/entity"
	"github.com/jp255ft-debug/cashout-engine/internal/core/port/inbound"
	"github.com/jp255ft-debug/cashout-engine/internal/core/port/outbound"
)

// ApproveTransferUseCase orquestra a aprovação de uma transferência.
type ApproveTransferUseCase struct {
	repo outbound.TransferRepository
}

// NewApproveTransferUseCase cria uma nova instância do caso de uso.
func NewApproveTransferUseCase(repo outbound.TransferRepository) *ApproveTransferUseCase {
	return &ApproveTransferUseCase{repo: repo}
}

// Execute executa o fluxo de aprovação de transferência.
func (uc *ApproveTransferUseCase) Execute(ctx context.Context, id, externalID string) (*inbound.TransferOutput, error) {
	// 1. Buscar transferência
	transfer, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("transferência não encontrada: %w", err)
	}

	// 2. Verificar se está pendente
	if transfer.Status() != entity.TransferStatusPending {
		return nil, fmt.Errorf("transferência não está pendente: status=%s", transfer.Status())
	}

	// 3. Aprovar no domínio
	if err := transfer.Approve(externalID); err != nil {
		return nil, fmt.Errorf("falha ao aprovar transferência: %w", err)
	}

	// 4. Persistir com optimistic locking
	if err := uc.repo.Update(ctx, transfer); err != nil {
		return nil, fmt.Errorf("falha ao atualizar transferência: %w", err)
	}

	return MapToTransferOutput(transfer), nil
}