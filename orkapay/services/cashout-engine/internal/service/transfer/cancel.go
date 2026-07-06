package transfer

import (
	"context"
	"fmt"

	"github.com/jp255ft-debug/cashout-engine/internal/core/port/inbound"
	"github.com/jp255ft-debug/cashout-engine/internal/core/port/outbound"
)

// CancelTransferUseCase orquestra o cancelamento de uma transferência.
type CancelTransferUseCase struct {
	repo     outbound.TransferRepository
	provider outbound.CashOutProviderGateway
}

// NewCancelTransferUseCase cria uma nova instância do caso de uso.
func NewCancelTransferUseCase(repo outbound.TransferRepository, provider outbound.CashOutProviderGateway) *CancelTransferUseCase {
	return &CancelTransferUseCase{repo: repo, provider: provider}
}

// Execute executa o fluxo de cancelamento de transferência.
func (uc *CancelTransferUseCase) Execute(ctx context.Context, id, reason string) (*inbound.TransferOutput, error) {
	// 1. Buscar transferência
	transfer, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("transferência não encontrada: %w", err)
	}

	// 2. Cancelar no provedor externo (se já tiver externalID)
	if transfer.ExternalID() != "" {
		if err := uc.provider.CancelTransfer(ctx, transfer.ExternalID()); err != nil {
			return nil, fmt.Errorf("falha ao cancelar transferência no provedor: %w", err)
		}
	}

	// 3. Cancelar no domínio
	if err := transfer.Cancel(reason); err != nil {
		return nil, fmt.Errorf("falha ao cancelar transferência: %w", err)
	}

	// 4. Persistir com optimistic locking
	if err := uc.repo.Update(ctx, transfer); err != nil {
		return nil, fmt.Errorf("falha ao atualizar transferência: %w", err)
	}

	return MapToTransferOutput(transfer), nil
}