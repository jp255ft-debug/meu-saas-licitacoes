package invoice

import (
	"context"
	"fmt"

	"github.com/jp255ft-debug/cashin-engine/internal/core/domain/entity"
	"github.com/jp255ft-debug/cashin-engine/internal/core/port/outbound"
)

// ListInvoicesByStatusUseCase lista faturas por status.
type ListInvoicesByStatusUseCase struct {
	repo outbound.InvoiceRepository
}

// NewListInvoicesByStatusUseCase cria uma nova instância do caso de uso.
func NewListInvoicesByStatusUseCase(repo outbound.InvoiceRepository) *ListInvoicesByStatusUseCase {
	return &ListInvoicesByStatusUseCase{repo: repo}
}

// Execute executa a listagem de faturas por status.
func (uc *ListInvoicesByStatusUseCase) Execute(ctx context.Context, status string, limit, offset int) ([]*entity.Invoice, error) {
	invoiceStatus := entity.InvoiceStatus(status)
	if invoiceStatus == "" {
		return nil, fmt.Errorf("status inválido: %s", status)
	}

	invoices, err := uc.repo.ListByStatus(ctx, invoiceStatus, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("falha ao listar faturas: %w", err)
	}
	return invoices, nil
}