package invoice

import (
	"context"
	"fmt"

	"github.com/jp255ft-debug/cashin-engine/internal/core/domain/entity"
	"github.com/jp255ft-debug/cashin-engine/internal/core/port/outbound"
)

// GetInvoiceByIDUseCase busca uma fatura pelo ID.
type GetInvoiceByIDUseCase struct {
	repo outbound.InvoiceRepository
}

// NewGetInvoiceByIDUseCase cria uma nova instância do caso de uso.
func NewGetInvoiceByIDUseCase(repo outbound.InvoiceRepository) *GetInvoiceByIDUseCase {
	return &GetInvoiceByIDUseCase{repo: repo}
}

// Execute executa a busca de fatura por ID.
func (uc *GetInvoiceByIDUseCase) Execute(ctx context.Context, id string) (*entity.Invoice, error) {
	invoice, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("fatura não encontrada: %w", err)
	}
	return invoice, nil
}