package invoice

import (
	"context"
	"fmt"

	"github.com/jp255ft-debug/cashin-engine/internal/core/port/inbound"
	"github.com/jp255ft-debug/cashin-engine/internal/core/port/outbound"
)

// CancelInvoiceUseCase orquestra o cancelamento de uma fatura.
type CancelInvoiceUseCase struct {
	repo     outbound.InvoiceRepository
	provider outbound.CashInProviderGateway
}

// NewCancelInvoiceUseCase cria uma nova instância do caso de uso.
func NewCancelInvoiceUseCase(repo outbound.InvoiceRepository, provider outbound.CashInProviderGateway) *CancelInvoiceUseCase {
	return &CancelInvoiceUseCase{repo: repo, provider: provider}
}

// Execute executa o fluxo de cancelamento de fatura.
func (uc *CancelInvoiceUseCase) Execute(ctx context.Context, id, reason string) (*inbound.InvoiceOutput, error) {
	// 1. Buscar fatura
	invoice, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("fatura não encontrada: %w", err)
	}

	// 2. Cancelar no provedor externo (se já tiver externalID)
	if invoice.ExternalID() != "" {
		if err := uc.provider.CancelInvoice(ctx, invoice.ExternalID()); err != nil {
			return nil, fmt.Errorf("falha ao cancelar fatura no provedor: %w", err)
		}
	}

	// 3. Cancelar no domínio (aplica regras de negócio)
	if err := invoice.Cancel(reason); err != nil {
		return nil, fmt.Errorf("falha ao cancelar fatura: %w", err)
	}

	// 4. Persistir com optimistic locking
	if err := uc.repo.Update(ctx, invoice); err != nil {
		return nil, fmt.Errorf("falha ao atualizar fatura: %w", err)
	}

	return MapToInvoiceOutput(invoice), nil
}