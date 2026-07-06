package invoice

import (
	"context"
	"fmt"

	"github.com/jp255ft-debug/cashin-engine/internal/core/domain/entity"
	"github.com/jp255ft-debug/cashin-engine/internal/core/port/inbound"
	"github.com/jp255ft-debug/cashin-engine/internal/core/port/outbound"
)

// IssueInvoiceUseCase orquestra a emissão de uma fatura pendente.
type IssueInvoiceUseCase struct {
	repo     outbound.InvoiceRepository
	provider outbound.CashInProviderGateway
}

// NewIssueInvoiceUseCase cria uma nova instância do caso de uso.
func NewIssueInvoiceUseCase(repo outbound.InvoiceRepository, provider outbound.CashInProviderGateway) *IssueInvoiceUseCase {
	return &IssueInvoiceUseCase{repo: repo, provider: provider}
}

// Execute executa o fluxo de emissão de fatura.
func (uc *IssueInvoiceUseCase) Execute(ctx context.Context, id string) (*inbound.InvoiceOutput, error) {
	// 1. Buscar fatura no repositório
	invoice, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("fatura não encontrada: %w", err)
	}

	// 2. Verificar se está pendente
	if invoice.Status() != entity.InvoiceStatusPending {
		return nil, fmt.Errorf("fatura não está pendente: status=%s", invoice.Status())
	}

	// 3. Criar no provedor externo
	externalID, pixQRCode, pixKey, err := uc.provider.CreateInvoice(ctx, invoice)
	if err != nil {
		return nil, fmt.Errorf("falha ao criar fatura no provedor: %w", err)
	}

	// 4. Emitir fatura
	if err := invoice.Issue(externalID, pixQRCode, pixKey); err != nil {
		return nil, fmt.Errorf("falha ao emitir fatura: %w", err)
	}

	// 5. Atualizar com optimistic locking
	if err := uc.repo.Update(ctx, invoice); err != nil {
		return nil, fmt.Errorf("falha ao atualizar fatura: %w", err)
	}

	return MapToInvoiceOutput(invoice), nil
}