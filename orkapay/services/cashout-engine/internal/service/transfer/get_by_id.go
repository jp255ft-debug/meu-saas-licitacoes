package transfer

import (
	"context"
	"fmt"

	"github.com/jp255ft-debug/cashout-engine/internal/core/domain/entity"
	"github.com/jp255ft-debug/cashout-engine/internal/core/port/outbound"
)

// GetTransferByIDUseCase busca uma transferência pelo ID.
type GetTransferByIDUseCase struct {
	repo outbound.TransferRepository
}

// NewGetTransferByIDUseCase cria uma nova instância do caso de uso.
func NewGetTransferByIDUseCase(repo outbound.TransferRepository) *GetTransferByIDUseCase {
	return &GetTransferByIDUseCase{repo: repo}
}

// Execute executa a busca de transferência por ID.
func (uc *GetTransferByIDUseCase) Execute(ctx context.Context, id string) (*entity.Transfer, error) {
	transfer, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("transferência não encontrada: %w", err)
	}
	return transfer, nil
}