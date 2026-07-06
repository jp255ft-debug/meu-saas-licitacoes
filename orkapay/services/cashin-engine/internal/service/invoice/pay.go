package invoice

import (
	"context"
	"fmt"

	"github.com/jp255ft-debug/cashin-engine/internal/core/port/inbound"
	"github.com/jp255ft-debug/cashin-engine/internal/core/port/outbound"
)

// PayInvoiceUseCase orquestra o registro de pagamento de uma fatura.
type PayInvoiceUseCase struct {
	repo outbound.InvoiceRepository
}

// NewPayInvoiceUseCase cria uma nova instância do caso de uso.
func NewPayInvoiceUseCase(repo outbound.InvoiceRepository) *PayInvoiceUseCase {
	return &PayInvoiceUseCase{repo: repo}
}

// Execute executa o fluxo de pagamento de fatura.
func (uc *PayInvoiceUseCase) Execute(ctx context.Context, id string, paidAmountCents int64, paymentMethod string) (*inbound.InvoiceOutput, error) {
	// 1. Buscar fatura
	invoice, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("fatura não encontrada: %w", err)
	}

	// 2. Registrar pagamento no domínio (aplica regras de negócio)
	if err := invoice.Pay(paidAmountCents, paymentMethod); err != nil {
		return nil, fmt.Errorf("falha ao registrar pagamento: %w", err)
	}

	// 3. Persistir atualização com optimistic locking
	if err := uc.repo.Update(ctx, invoice); err != nil {
		return nil, fmt.Errorf("falha ao atualizar fatura: %w", err)
	}

	return MapToInvoiceOutput(invoice), nil
}